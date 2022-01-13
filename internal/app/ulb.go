package app

import (
	"fmt"
	"github.com/ucloud/migrate/internal/log"
	"time"
)

func (app *MigrateApp) MigrateULB() error {
	// Start Get Cube IP List
	log.Logger.Sugar().Infof("[Start Get ULB List with Cube]")
	ulbCubeInfos, err := app.MigrateService.GetULBCubeInfoList(app.Config.MigrateULB.ULBConfig)
	if err != nil {
		return fmt.Errorf("[Get ULB List with Cube] got error, %s", err)
	}

	if len(ulbCubeInfos) == 0 {
		return fmt.Errorf("[Get ULB List with Cube] got empty ULB List with Cube Backend")
	}

	//log.Logger.Sugar().Infof("[Get Cube List binded to ULB %s]", ulbCubeInfo.ULBId)
	cubeIdMap := make(map[string]struct{})
	cubeIdList := make([]string, 0)
	for _, ulbCubeInfo := range ulbCubeInfos {
		for _, vServerInfo := range ulbCubeInfo.VServerCubeInfos {
			for cubeId := range vServerInfo.CubeBackendMap {
				if _, ok := cubeIdMap[cubeId]; !ok {
					cubeIdMap[cubeId] = struct{}{}
					cubeIdList = append(cubeIdList, cubeId)
				}
			}
		}
	}

	// Start CreateUHost
	log.Logger.Sugar().Infof("[Start CreateUHost] for the Cube List %v", cubeIdList)
	var uHostIds []string
	uHostIds, err = app.MigrateService.CreateUHost(app.Config.MigrateULB.UHostConfig, len(cubeIdList))
	if err != nil {
		return fmt.Errorf("[CreateUHost] got error, %s", err)
	}
	if len(uHostIds) == 0 {
		return fmt.Errorf("[CreateUHost] got no one uhost can be created about the config, %v", *app.Config.MigrateULB.UHostConfig)
	}
	if len(uHostIds) < len(cubeIdList) {
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
		return fmt.Errorf("[Waiting for UHost Running] got no one uhost to be Running about the config, %v", app.Config.MigrateULB.UHostConfig)
	}

	// wait for service start on UHost
	time.Sleep(10 * time.Second)

	// Start Migrate
	log.Logger.Sugar().Infof("[Start Migrate], %v", runningUHostIds)

	migrateUHostIds := runningUHostIds
	successfulCubeBackends := make([]string, 0)
	unSuccessfulCubeBackends := make([]string, 0)
	for _, ulbCubeInfo := range ulbCubeInfos {
		for _, vServerInfo := range ulbCubeInfo.VServerCubeInfos {
			//vServerId := vServerInfo.VServerId
			//policies  := vServerInfo.Policies
			for cubeId, backendIds := range vServerInfo.CubeBackendMap {
				var uhostId string
				if len(migrateUHostIds) > 0 {
					uhostId = migrateUHostIds[len(migrateUHostIds)-1]
					migrateUHostIds = migrateUHostIds[:len(migrateUHostIds)-1]
				} else {
					log.Logger.Sugar().Warnf("[Start Migrate] not have enough running UHost for migrate: "+
						"ULBId[%s]:VServerId[%s]:CubeId[%s]:BackendIds[%v]", ulbCubeInfo.ULBId, vServerInfo.VServerId, cubeId, backendIds)
					return nil
				}

				for _, backend := range backendIds {
					var newBackendId string
					// BindUHostToUlBVServer
					log.Logger.Sugar().Infof("[BindUHostToUlBVServer] about "+
						"ULBId[%s]:VServerId[%s]:BackendId[%s]:CubeId[%s]UHostId[%s]",
						ulbCubeInfo.ULBId, vServerInfo.VServerId, backend.BackendId, cubeId, uhostId)
					newBackendId, err = app.MigrateService.BindUHostToUlBVServer(ulbCubeInfo.ULBId, vServerInfo.VServerId, uhostId, backend.Port)
					if err != nil {
						log.Logger.Sugar().Warnf("[BindUHostToUlBVServer] about "+
							"ULBId[%s]:VServerId[%s]:BackendId[%s]:CubeId[%s]UHostId[%s] got error ,%s",
							ulbCubeInfo.ULBId, vServerInfo.VServerId, backend.BackendId, cubeId, uhostId, err)

						unSuccessfulCubeBackends = append(unSuccessfulCubeBackends,
							fmt.Sprintf("ULBId[%s]:VServerId[%s]:BackendId[%s]:CubeId[%s]:UHostId[%s]",
								ulbCubeInfo.ULBId, vServerInfo.VServerId, backend.BackendId, cubeId, uhostId))
						continue
					}
					//todo validate and migrate policy

					// UnBindBackendToUlB
					log.Logger.Sugar().Infof("[UnBindBackendToUlB] about "+
						"ULBId[%s]:VServerId[%s]:BackendId[%s]:CubeId[%s]:UHostId[%s]",
						ulbCubeInfo.ULBId, vServerInfo.VServerId, backend.BackendId, cubeId, uhostId)
					if err = app.MigrateService.UnBindBackendToUlB(ulbCubeInfo.ULBId, backend.BackendId); err != nil {
						unSuccessfulCubeBackends = append(unSuccessfulCubeBackends,
							fmt.Sprintf("ULBId[%s]:VServerId[%s]:BackendId[%s]:CubeId[%s]:UHostId[%s]",
								ulbCubeInfo.ULBId, vServerInfo.VServerId, backend.BackendId, cubeId, uhostId))

						log.Logger.Sugar().Warnf("[UnBindBackendToUlB] about "+
							"ULBId[%s]:VServerId[%s]:BackendId[%s]:CubeId[%s]:UHostId[%s] got error ,%s",
							ulbCubeInfo.ULBId, vServerInfo.VServerId, backend.BackendId, cubeId, uhostId, err)

						log.Logger.Sugar().Infof("[Rollback - UnBindUHostToUlB] about "+
							"ULBId[%s]:VServerId[%s]:NewBackendId[%s]:CubeId[%s]:UHostId[%s]",
							ulbCubeInfo.ULBId, vServerInfo.VServerId, newBackendId, cubeId, uhostId)
						if err = app.MigrateService.UnBindBackendToUlB(ulbCubeInfo.ULBId, newBackendId); err != nil {
							log.Logger.Sugar().Warnf("[Rollback - UnBindUHostToUlB] about "+
								"ULBId[%s]:VServerId[%s]:NewBackendId[%s]:CubeId[%s]:UHostId[%s] got error ,%s",
								ulbCubeInfo.ULBId, vServerInfo.VServerId, newBackendId, cubeId, uhostId, err)
						}
						continue
					}
					log.Logger.Sugar().Infof("[Migrate ULB Backend Successful] about "+
						"ULBId[%s]:VServerId[%s]:BackendId[%s]:CubeId[%s]:UHostId[%s]",
						ulbCubeInfo.ULBId, vServerInfo.VServerId, backend.BackendId, cubeId, uhostId)
					successfulCubeBackends = append(successfulCubeBackends,
						fmt.Sprintf("ULBId[%s]:VServerId[%s]:BackendId[%s]:CubeId[%s]:UHostId[%s]",
							ulbCubeInfo.ULBId, vServerInfo.VServerId, backend.BackendId, cubeId, uhostId))
				}
			}
		}
	}
	log.Logger.Sugar().Infof("[All Migrate Finished]"+
		"[Successful] ULB Backend List: (%d)%v,"+
		"[Unsuccessful] ULB Backend List: (%d)%v",
		len(successfulCubeBackends), successfulCubeBackends, len(unSuccessfulCubeBackends), unSuccessfulCubeBackends)
	return nil
}
