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
	"net"
	"strconv"
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

	type ipInfo struct {
		ip    string
		eipId string
	}
	cubeIPMap := make(map[string][]ipInfo)
	cubeIdList := make([]string, 0)
	allIpList := make([]string, 0)
	for _, info := range cubeInfos {
		ipList := make([]string, 0)
		ipInfoList := make([]ipInfo, 0)
		if info.Eip != nil && len(info.Eip) > 0 {
			for _, eip := range info.Eip {
				ipList = append(ipList, eip.EIPAddr[0].IP)
				ipInfoList = append(ipInfoList, ipInfo{
					ip:    eip.EIPAddr[0].IP,
					eipId: eip.EIPId,
				})
			}
			cubeIdList = append(cubeIdList, info.CubeId)
			cubeIPMap[info.CubeId] = ipInfoList
			allIpList = append(allIpList, ipList...)
		}
	}
	if len(cubeIPMap) == 0 {
		return fmt.Errorf("got empty Cube list with external IP")
	}

	log.Logger.Sugar().Infof("[Start CreateUHost] for the Cube external IP List %v", allIpList)
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

				lock.Lock()
				allUHostIds = append(allUHostIds, ids...)
				lock.Unlock()
			}(*app.Config.Migrate, maxCount)
		}
		wg.Wait()
	} else {
		allUHostIds, err = migrateService.CreateUHost(app.Config.Migrate.UHostConfig, len(cubeIPMap))
		if err != nil {
			return fmt.Errorf("[CreateUHost] got error, %s", err)
		}
	}

	if len(allUHostIds) == 0 {
		return fmt.Errorf("[CreateUHost] got no one uhost can be created about the config, %v", *app.Config.Migrate.UHostConfig)
	}
	if len(allUHostIds) < len(cubeIPMap) {
		log.Logger.Sugar().Warnf("[CreateUHost] not created expected count UHost, " +
			"please check the specific UHost quota")
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

	// wait for service start on UHost
	time.Sleep(10 * time.Second)

	log.Logger.Sugar().Infof("[Start Migrate], %v", availableUHostIds)
	successfulIps := make([]string, 0)
	for i := 0; i < len(availableUHostIds); i++ {
		for _, info := range cubeIPMap[cubeIdList[i]] {
			log.Logger.Sugar().Infof("[UnBindEIPToCube] about cubeId[%s] and ip[%s:%s]", cubeIdList[i], info.eipId, info.ip)
			if err = migrateService.UnBindEIPToCube(cubeIdList[i], info.eipId); err != nil {
				log.Logger.Sugar().Warnf("[UnBindEIPToCube] about cubeId[%s] and ip[%s:%s] got error, %s", cubeIdList[i], info.eipId, info.ip, err)
				continue
			}

			log.Logger.Sugar().Infof("[BindEIPToUHost] about uhostId[%s] and ip[%s:%s]", availableUHostIds[i], info.eipId, info.ip)
			if err = migrateService.BindEIPToUHost(availableUHostIds[i], info.eipId); err != nil {
				log.Logger.Sugar().Warnf("[BindEIPToUHost] about uhostId[%s] and ip[%s:%s] got error, %s", availableUHostIds[i], info.eipId, info.ip, err)

				log.Logger.Sugar().Infof("[ReBindEIPToCube] about cubeId[%s] and ip[%s:%s]", cubeIdList[i], info.eipId, info.ip)
				if err = migrateService.BindEIPToCube(cubeIdList[i], info.eipId); err != nil {
					return fmt.Errorf("[ReBindEIPToCube] about cubeId[%s] and ip[%s:%s] got error, %s", cubeIdList[i], info.eipId, info.ip, err)
				}
			} else {
				if app.Config.Migrate.ServiceValidation != nil && app.Config.Migrate.ServiceValidation.Port != 0 {
					log.Logger.Sugar().Infof("[Waiting For UHost Server Running] about uhostId[%s], ip[%s:%s], port[%d]",
						availableUHostIds[i], info.eipId, info.ip, app.Config.Migrate.ServiceValidation.Port)

					timeout := app.Config.Migrate.ServiceValidation.WaitServiceReadyTimeout
					if timeout == 0 {
						timeout = 120
					}
					err = retry.Config{
						StartTimeout: time.Duration(timeout) * time.Second,
						ShouldRetry: func(err error) bool {
							return errors.IsNotCompleteError(err)
						},
						RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 2}).Linear,
					}.Run(ctx, func(ctx context.Context) error {
						var conn net.Conn
						conn, err = net.DialTimeout("tcp", info.ip+":"+strconv.Itoa(app.Config.Migrate.ServiceValidation.Port), 3*time.Second)
						if err != nil {
							return errors.NewNotCompletedError(err)
						}

						return conn.Close()
					})

					if err != nil {
						log.Logger.Sugar().Warnf("[Waiting For UHost Server Port Running] about uhostId[%s], ip[%s:%s], port[%d] got error, %s",
							availableUHostIds[i], info.eipId, info.ip, app.Config.Migrate.ServiceValidation.Port, err)

						log.Logger.Sugar().Infof("[UnBindEIPToUHost] about uhostId[%s] and ip[%s:%s]", availableUHostIds[i], info.eipId, info.ip)
						if err = migrateService.UnBindEIPToUHost(availableUHostIds[i], info.eipId); err != nil {
							return fmt.Errorf("[UnBindEIPToUHost] about uhostId[%s] and ip[%s:%s] got error, %s", availableUHostIds[i], info.eipId, info.ip, err)
						}

						log.Logger.Sugar().Infof("[ReBindEIPToCube] about cubeId[%s] and ip[%s:%s]", cubeIdList[i], info.eipId, info.ip)
						if err = migrateService.BindEIPToCube(cubeIdList[i], info.eipId); err != nil {
							return fmt.Errorf("[ReBindEIPToCube] about cubeId[%s] and ip[%s:%s] got error, %s", cubeIdList[i], info.eipId, info.ip, err)
						}
					}
				}

				log.Logger.Sugar().Infof("[Migrate One IP Successful] about uhostId[%s] and ip[%s:%s]", availableUHostIds[i], info.eipId, info.ip)
				successfulIps = append(successfulIps, info.ip)
			}
		}
	}

	differSlice := utils.DifferenceSlice(allIpList, successfulIps)
	log.Logger.Sugar().Infof("[All Migrate Successful] Want Migrate the EIP List: (%d)%v,"+
		"[Successful] about external IP List: (%d)%v,"+
		"[Unsuccessful] external IP List: (%d)%v",
		len(allIpList), allIpList, len(successfulIps), successfulIps, len(differSlice), differSlice)

	return nil

}
