package app

import (
	"context"
	"fmt"
	"github.com/ucloud/migrate/internal/errors"
	"github.com/ucloud/migrate/internal/log"
	"github.com/ucloud/migrate/internal/utils"
	"github.com/ucloud/migrate/internal/utils/retry"
	"net"
	"strconv"
	"time"
)

func (app *MigrateApp) MigrateEIP() error {
	// Start Get Cube IP List
	log.Logger.Sugar().Infof("[Start Get Cube EIP(ExtranetIP) List]")
	cubeInfos, err := app.MigrateService.GetCubePodExtendInfoList(app.Config.MigrateEIP.CubeConfig)
	if err != nil {
		return fmt.Errorf("[Get Cube EIP(ExtranetIP) List] got error, %s", err)
	}
	type ipInfo struct {
		ip    string
		eipId string
	}
	// Cube ID to IP List
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
		return fmt.Errorf("[Get Cube EIP(ExtranetIP) List] got empty Cube list with EIP(ExtranetIP)")
	}

	// Start CreateUHost
	log.Logger.Sugar().Infof("[Start CreateUHost] for the Cube EIP(ExtranetIP) List %v", allIpList)
	var uHostIds []string
	uHostIds, err = app.MigrateService.CreateUHost(app.Config.MigrateEIP.UHostConfig, len(cubeIPMap))
	if err != nil {
		return fmt.Errorf("[CreateUHost] got error, %s", err)
	}
	if len(uHostIds) == 0 {
		return fmt.Errorf("[CreateUHost] got no one uhost can be created about the config, %v", *app.Config.MigrateEIP.UHostConfig)
	}
	if len(uHostIds) < len(cubeIPMap) {
		log.Logger.Sugar().Warnf("[CreateUHost] not created expected count UHost, " +
			"please check the specific UHost quota or instacne type not enough")
	}

	// Waiting for UHost Running
	log.Logger.Sugar().Infof("[Waiting for UHost Running], %v", uHostIds)
	var runningUHostIds []string
	runningUHostIds, err = app.MigrateService.WaitingForUHostRunning(uHostIds)
	if err != nil {
		return fmt.Errorf("[Waiting for UHost Running] got error, %s", err)
	}
	if len(runningUHostIds) == 0 {
		return fmt.Errorf("[Waiting for UHost Running] got no one uhost to be Running about the config, %v", app.Config.MigrateEIP.UHostConfig)
	}

	// wait for service start on UHost
	time.Sleep(10 * time.Second)

	// Start Migrate
	log.Logger.Sugar().Infof("[Start Migrate], %v", runningUHostIds)
	successfulIps := make([]string, 0)
	for i := 0; i < len(runningUHostIds); i++ {
		for _, info := range cubeIPMap[cubeIdList[i]] {
			log.Logger.Sugar().Infof("[UnBindEIPToCube] about cubeId[%s] and ip[%s:%s]", cubeIdList[i], info.eipId, info.ip)
			if err = app.MigrateService.UnBindEIPToCube(cubeIdList[i], info.eipId); err != nil {
				log.Logger.Sugar().Warnf("[UnBindEIPToCube] about cubeId[%s] and ip[%s:%s] got error, %s", cubeIdList[i], info.eipId, info.ip, err)
				continue
			}

			log.Logger.Sugar().Infof("[BindEIPToUHost] about uhostId[%s] and ip[%s:%s]", runningUHostIds[i], info.eipId, info.ip)
			if err = app.MigrateService.BindEIPToUHost(runningUHostIds[i], info.eipId); err != nil {
				log.Logger.Sugar().Warnf("[BindEIPToUHost] about uhostId[%s] and ip[%s:%s] got error, %s", runningUHostIds[i], info.eipId, info.ip, err)

				log.Logger.Sugar().Infof("[Rollback - ReBindEIPToCube] about cubeId[%s] and ip[%s:%s]", cubeIdList[i], info.eipId, info.ip)
				if err = app.MigrateService.BindEIPToCube(cubeIdList[i], info.eipId); err != nil {
					return fmt.Errorf("[Rollback - ReBindEIPToCube] about cubeId[%s] and ip[%s:%s] got error, %s", cubeIdList[i], info.eipId, info.ip, err)
				}
			} else {
				if app.Config.MigrateEIP.ServiceValidation != nil && app.Config.MigrateEIP.ServiceValidation.Port != 0 {
					log.Logger.Sugar().Infof("[Waiting For UHost Server Running] about uhostId[%s], ip[%s:%s], port[%d]",
						runningUHostIds[i], info.eipId, info.ip, app.Config.MigrateEIP.ServiceValidation.Port)

					timeout := app.Config.MigrateEIP.ServiceValidation.WaitServiceReadyTimeout
					if timeout == 0 {
						timeout = 120
					}
					ctx := context.TODO()
					err = retry.Config{
						StartTimeout: time.Duration(timeout) * time.Second,
						ShouldRetry: func(err error) bool {
							return errors.IsNotCompleteError(err)
						},
						RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 2}).Linear,
					}.Run(ctx, func(ctx context.Context) error {
						var conn net.Conn
						conn, err = net.DialTimeout("tcp", info.ip+":"+strconv.Itoa(app.Config.MigrateEIP.ServiceValidation.Port), 3*time.Second)
						if err != nil {
							return errors.NewNotCompletedError(err)
						}

						return conn.Close()
					})

					if err != nil {
						log.Logger.Sugar().Warnf("[Waiting For UHost Server Port Running] about uhostId[%s], ip[%s:%s], port[%d] got error, %s",
							runningUHostIds[i], info.eipId, info.ip, app.Config.MigrateEIP.ServiceValidation.Port, err)

						log.Logger.Sugar().Infof("[Rollback - UnBindEIPToUHost] about uhostId[%s] and ip[%s:%s]", runningUHostIds[i], info.eipId, info.ip)
						if err = app.MigrateService.UnBindEIPToUHost(runningUHostIds[i], info.eipId); err != nil {
							// todo test logger error to replace return error
							return fmt.Errorf("[Rollback - UnBindEIPToUHost] about uhostId[%s] and ip[%s:%s] got error, %s", runningUHostIds[i], info.eipId, info.ip, err)
						}

						log.Logger.Sugar().Infof("[Rollback - ReBindEIPToCube] about cubeId[%s] and ip[%s:%s]", cubeIdList[i], info.eipId, info.ip)
						if err = app.MigrateService.BindEIPToCube(cubeIdList[i], info.eipId); err != nil {
							return fmt.Errorf("[Rollback - ReBindEIPToCube] about cubeId[%s] and ip[%s:%s] got error, %s", cubeIdList[i], info.eipId, info.ip, err)
						}
					}
				}

				log.Logger.Sugar().Infof("[Migrate One IP Successful] about uhostId[%s] and ip[%s:%s]", runningUHostIds[i], info.eipId, info.ip)
				successfulIps = append(successfulIps, info.ip)
			}
		}
	}

	differSlice := utils.DifferenceSlice(allIpList, successfulIps)
	log.Logger.Sugar().Infof("[All Migrate Finished]"+
		"[Successful] EIP(ExtranetIP) List: (%d)%v,"+
		"[Unsuccessful] EIP(ExtranetIP) List: (%d)%v",
		len(successfulIps), successfulIps, len(differSlice), differSlice)

	return nil
}
