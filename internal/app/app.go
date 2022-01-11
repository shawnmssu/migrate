package app

import (
	"context"
	"fmt"
	"github.com/ucloud/migrate/internal/client"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/migrate/internal/errors"
	"github.com/ucloud/migrate/internal/log"
	"github.com/ucloud/migrate/internal/service"
	"github.com/ucloud/migrate/internal/utils"
	"github.com/ucloud/migrate/internal/utils/retry"
	"strings"
	"sync"
	"time"
)

type MigrateApp struct {
	Config conf.Config
}

func NewMigrateApp(bc conf.Config) *MigrateApp {
	return &MigrateApp{
		bc,
	}
}

func (app *MigrateApp) Run() {
	globalClient, err := client.NewClient(app.Config)
	if err != nil {
		log.Logger.Sugar().Errorf("New SDK Clinet Got error, %s", err)
		return
	}

	migrateService := service.NewBasicMigrateCubeService(globalClient)
	if err = app.migrate(migrateService); err != nil {
		log.Logger.Sugar().Errorf("Migrate Got error, %s", err)
		return
	}
}

func (app *MigrateApp) migrate(migrateService service.MigrateCubeService) error {
	log.Logger.Sugar().Infof("[Start Get Cube IP List]")
	cubeInfos, err := migrateService.GetCubeCubePodExtendInfoList(app.Config.Migrate.CubeConfig)
	if err != nil {
		return err
	}

	cubeIPMap := make(map[string][]string)
	cubeIdList := make([]string, 0)
	allEipList := make([]string, 0)
	for _, info := range cubeInfos {
		eipIdList := make([]string, 0)
		if info.Eip != nil && len(info.Eip) > 0 {
			for _, eip := range info.Eip {
				eipIdList = append(eipIdList, eip.EIPId)
			}
			cubeIdList = append(cubeIdList, info.CubeId)
			cubeIPMap[info.CubeId] = eipIdList
			allEipList = append(allEipList, eipIdList...)
		}
	}

	log.Logger.Sugar().Infof("[Start CreateUHost] for the Cube EIP List %v", allEipList)
	var allUHostIds []string
	count := len(cubeIPMap) / 10
	if count > 1 {
		var wg sync.WaitGroup
		var lock sync.Mutex
		for i := 1; i <= len(cubeIPMap)/10; i++ {
			maxCount := 10
			if i == len(cubeIPMap)/10 && len(cubeIPMap)%10 > 0 {
				maxCount = 10 + len(cubeIPMap)%10
			}
			wg.Add(1)
			go func(migrateConf conf.Migrate, maxCount int) {
				defer wg.Done()
				var ids []string
				ids, err = migrateService.CreateUHost(migrateConf.UHostConfig, maxCount)
				if err != nil {
					log.Logger.Sugar().Warnf("[CreateUHost] got error, %s", err)
					return
				}
				if len(ids) < maxCount {
					log.Logger.Sugar().Warnf("[CreateUHost] not created expected count UHost, " +
						"please check the specific UHost quota")
					return
				}

				lock.Lock()
				allUHostIds = append(allUHostIds, ids...)
				lock.Unlock()
			}(*app.Config.Migrate, maxCount)
		}
		wg.Wait()
	} else {
		allUHostIds, err = migrateService.CreateUHost(app.Config.Migrate.UHostConfig, len(cubeIPMap))
		if err != nil {
			return err
		}
	}

	if len(allUHostIds) == 0 {
		return fmt.Errorf("[CreateUHost] got no one uhost can be created about the config, %v", *app.Config.Migrate.UHostConfig)
	}

	log.Logger.Sugar().Infof("[Waiting for UHost Running], %v", allUHostIds)

	expectedUHostIds := map[string]struct{}{}
	for _, id := range allUHostIds {
		expectedUHostIds[id] = struct{}{}
	}
	availableUHostIds := make([]string, 0)
	ctx := context.TODO()
	err = retry.Config{
		Tries: 10,
		ShouldRetry: func(err error) bool {
			return errors.IsNotCompleteError(err)
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 10}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		for uhostId := range expectedUHostIds {
			inst, err := migrateService.DescribeUHostById(uhostId)
			if err != nil {
				return err
			}

			if inst.State == "Running" {
				delete(expectedUHostIds, uhostId)
				availableUHostIds = append(availableUHostIds, uhostId)
				continue
			}

			if inst.State == "Install Fail" {
				delete(expectedUHostIds, uhostId)
				log.Logger.Sugar().Warnf("Install UHost %s failed", uhostId)
				continue
			}
		}

		if len(expectedUHostIds) != 0 {
			keys := make([]string, len(expectedUHostIds))
			for k := range expectedUHostIds {
				keys = append(keys, k)
			}
			return errors.NewNotCompletedError(fmt.Errorf("waiting uhost %v running", keys))
		}

		return nil
	})

	if err != nil {
		var s []string
		for v := range expectedUHostIds {
			s = append(s, v)
		}
		return fmt.Errorf("[Waiting for UHost Running] about uhostIds: %q got error, %s", strings.Join(s, ","), err)
	}

	if len(availableUHostIds) == 0 {
		return fmt.Errorf("[Waiting for UHost Running] got no one uhost to be Running about the config, %v", app.Config.Migrate.UHostConfig)
	}

	log.Logger.Sugar().Infof("[Start Migrate], %v", availableUHostIds)
	successfulIps := make([]string, 0)
	for i := 0; i < len(availableUHostIds); i++ {
		for _, ip := range cubeIPMap[cubeIdList[i]] {
			if err = migrateService.UnBindEIPToCube(cubeIdList[i], ip); err != nil {
				log.Logger.Sugar().Warnf("[UnBindEIPToCube] about cubeId[%s] and ip[%s] got error, %s", cubeIdList[i], ip, err)
				continue
			}

			log.Logger.Sugar().Infof("[UnBindEIPToCube] about cubeId[%s] and ip[%s] complete", cubeIdList[i], ip)

			if err = migrateService.BindEIPToUHost(availableUHostIds[i], ip); err != nil {
				log.Logger.Sugar().Warnf("[BindEIPToUHost] about uhostId[%s] and ip[%s] got error, %s", availableUHostIds[i], ip, err)
				if err = migrateService.BindEIPToCube(cubeIdList[i], ip); err != nil {
					log.Logger.Sugar().Errorf("[BindEIPToCube] about cubeId[%s] and ip[%s] got error, %s", cubeIdList[i], ip, err)
					return fmt.Errorf("[BindEIPToCube] about cubeId[%s] and ip[%s] got error, %s", cubeIdList[i], ip, err)
				} else {
					log.Logger.Sugar().Infof("[ReBindEIPToCube] about cubeId[%s] and ip[%s] complete", cubeIdList[i], ip)
				}
			} else {
				log.Logger.Sugar().Infof("[BindEIPToUHost] about uhostId[%s] and ip[%s] complete", availableUHostIds[i], ip)
				successfulIps = append(successfulIps, ip)
			}
		}
	}

	differSlice := utils.DifferenceSlice(allEipList, successfulIps)
	log.Logger.Sugar().Infof("[Successful Migrate] Want Migrate the EIP List: [%d]%v, "+
		"Successful about EIP List: [%d]%v, "+
		"Unsuccessful EIP List: [%d]%v",
		len(allEipList), allEipList, len(successfulIps), successfulIps, len(differSlice), differSlice)

	return nil

}
