package app

import (
	"fmt"
	"github.com/ucloud/migrate/internal/log"
	"time"
)

func (app *MigrateApp) MigrateULB() error {
	// Start Get Cube IP List
	log.Logger.Sugar().Infof("[Start Get ULB VServer List with Cube]")
	vServerCubeInfos, err := app.MigrateService.GetULBVServerInfoListAboutCube(app.Config.MigrateULB.ULBId)
	if err != nil {
		return fmt.Errorf("[Get ULB VServer List with Cube] got error, %s", err)
	}

	if len(vServerCubeInfos) == 0 {
		return fmt.Errorf("[Get ULB VServer List with Cube] got empty ULB VServer List with Cube Backend")
	}

	cubeIdMap := make(map[string]struct{})
	cubeIdList := make([]string, 0)
	for _, vServerInfo := range vServerCubeInfos {
		for cubeId := range vServerInfo.CubeBackendMap {
			if _, ok := cubeIdMap[cubeId]; !ok {
				cubeIdMap[cubeId] = struct{}{}
				cubeIdList = append(cubeIdList, cubeId)
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
	cubeUHostIdMap := make(map[string]string)

	var notEnoughUHost bool
	for _, vServerInfo := range vServerCubeInfos {
		for cubeId, backendIds := range vServerInfo.CubeBackendMap {
			var uhostId string
			if _, ok := cubeUHostIdMap[cubeId]; ok {
				uhostId = cubeUHostIdMap[cubeId]
			} else if len(migrateUHostIds) > 0 {
				uhostId = migrateUHostIds[len(migrateUHostIds)-1]
				migrateUHostIds = migrateUHostIds[:len(migrateUHostIds)-1]
				cubeUHostIdMap[cubeId] = uhostId
			} else {
				if !notEnoughUHost {
					log.Logger.Sugar().Warnf("[Start Migrate] not have enough running UHost for migrate: "+
						"VServerId(%s):CubeId(%s):BackendIds[%v]", vServerInfo.VServerId, cubeId, backendIds)
					break
				}
				notEnoughUHost = true
				continue
			}

			for _, backend := range backendIds {
				// BindUHostToUlBVServer
				log.Logger.Sugar().Infof("[BindUHostToUlBVServer] about "+
					"VServerId(%s):BackendId(%s):CubeId(%s)UHostId(%s)",
					vServerInfo.VServerId, backend.BackendId, cubeId, uhostId)
				newBackendId, err := app.MigrateService.BindUHostToUlBVServer(app.Config.MigrateULB.ULBId, vServerInfo.VServerId, uhostId, backend.Port)
				if err != nil {
					return fmt.Errorf("[BindUHostToUlBVServer] about "+
						"VServerId(%s):BackendId(%s):CubeId(%s)UHostId(%s) got error ,%s",
						vServerInfo.VServerId, backend.BackendId, cubeId, uhostId, err)
				}
				//todo validate and migrate policy

				// UnBindBackendToUlB
				log.Logger.Sugar().Infof("[UnBindBackendToUlB] about "+
					"VServerId(%s):BackendId(%s):CubeId(%s):UHostId(%s)",
					vServerInfo.VServerId, backend.BackendId, cubeId, uhostId)
				if err = app.MigrateService.UnBindBackendToUlB(app.Config.MigrateULB.ULBId, backend.BackendId); err != nil {
					log.Logger.Sugar().Warnf("[UnBindBackendToUlB] about "+
						"VServerId(%s):BackendId(%s):CubeId(%s):UHostId(%s) got error ,%s",
						vServerInfo.VServerId, backend.BackendId, cubeId, uhostId, err)

					log.Logger.Sugar().Infof("[Rollback - UnBindUHostToUlB] about "+
						"VServerId(%s):NewBackendId(%s):CubeId(%s):UHostId(%s)",
						vServerInfo.VServerId, newBackendId, cubeId, uhostId)
					if err = app.MigrateService.UnBindBackendToUlB(app.Config.MigrateULB.ULBId, newBackendId); err != nil {
						log.Logger.Sugar().Warnf("[Rollback - UnBindUHostToUlB] about "+
							"VServerId(%s):NewBackendId(%s):CubeId(%s):UHostId(%s) got error ,%s",
							vServerInfo.VServerId, newBackendId, cubeId, uhostId, err)
					}
					return fmt.Errorf("[UnBindBackendToUlB] about "+
						"VServerId(%s):BackendId(%s):CubeId(%s):UHostId(%s) got error ,%s",
						vServerInfo.VServerId, backend.BackendId, cubeId, uhostId, err)
				}
				log.Logger.Sugar().Infof("[Migrate ULB Backend Successful] about "+
					"VServerId(%s):BackendId(%s):CubeId(%s):UHostId(%s)",
					vServerInfo.VServerId, backend.BackendId, cubeId, uhostId)
				successfulCubeBackends = append(successfulCubeBackends,
					fmt.Sprintf("VServerId(%s):BackendId(%s):CubeId(%s):UHostId(%s)",
						vServerInfo.VServerId, backend.BackendId, cubeId, uhostId))
			}
		}
	}

	log.Logger.Sugar().Infof("[All Migrate Finished about ULBId(%s)]"+
		"[Successful] ULB Backend List: (%d)%v,"+
		"[Need CLean UP] UHost List: (%d)%v",
		app.Config.MigrateULB.ULBId, len(successfulCubeBackends), successfulCubeBackends,
		len(migrateUHostIds), migrateUHostIds)

	return nil
}
