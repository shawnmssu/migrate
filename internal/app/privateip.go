package app

import (
	"fmt"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/migrate/internal/log"
	"time"
)

func (app *MigrateApp) MigratePrivateIp(dryRun bool) error {
	// Start Get Cube PrivateIp List
	log.Logger.Sugar().Infof("[Start Get Cube PrivateIp List]")
	cubeInfos, err := app.MigrateService.GetCubePodInfoList(app.Config.MigratePrivateIp.CubeConfig)
	if err != nil {
		return fmt.Errorf("[Get Cube PrivateIp List] got error, %s", err)
	}
	type cubePrivateIpInfo struct {
		cubeId    string
		privateIp string
		vpcId     string
		subnetId  string
	}

	cubePrivateIpInfoList := make([]cubePrivateIpInfo, 0)
	for _, info := range cubeInfos {
		cubePrivateIpInfoList = append(cubePrivateIpInfoList, cubePrivateIpInfo{
			cubeId:    info.Metadata.CubeId,
			privateIp: info.Status.PodIp,
			vpcId:     info.Metadata.Provider.VpcId,
			subnetId:  info.Metadata.Provider.SubnetId,
		})
	}
	log.Logger.Sugar().Infof("[Need Migrate Cube List], %v", cubePrivateIpInfoList)

	if dryRun {
		// Try to CreateUHost
		log.Logger.Sugar().Infof("[Try to Create One UHost] about vpcId(%s) and subnetId(%s)",
			cubePrivateIpInfoList[0].vpcId, cubePrivateIpInfoList[0].subnetId)
		if err = app.tryCreateUHost(*app.Config.MigratePrivateIp.UHostConfig,
			cubePrivateIpInfoList[0].vpcId, cubePrivateIpInfoList[0].subnetId); err != nil {
			return err
		}
		log.Logger.Sugar().Infof("[Dry Run Complete]")
		return nil
	}

	successfulMigrateUHosts := make([]string, 0)
	// Start Migrate
	log.Logger.Sugar().Infof("[Start Migrate], %v", cubePrivateIpInfoList)
	for _, info := range cubePrivateIpInfoList {
		log.Logger.Sugar().Infof("[Start Migrate Cube %s] with privateIp(%s)", info.cubeId, info.privateIp)
		// Delete Cube
		log.Logger.Sugar().Infof("[Delete Cube %s] with privateIp(%s)", info.cubeId, info.privateIp)
		if err = app.MigrateService.DeleteCube(info.cubeId); err != nil {
			return fmt.Errorf("[Delete Cube %s] with privateIp(%s) got error, %s", info.cubeId, info.privateIp, err)
		}

		// wait for privateIp unbind to Cube
		log.Logger.Sugar().Infof("[Wait for privateIp %s unbind to Cube %s]", info.privateIp, info.cubeId)
		time.Sleep(15 * time.Second)

		// Start CreateUHost
		log.Logger.Sugar().Infof("[Start CreateUHost] about vpcId(%s):subnetId(%s):CubeId(%s):privateIp(%s)",
			info.vpcId, info.subnetId, info.cubeId, info.privateIp)
		uhostConfigCopy := *app.Config.MigrateEIP.UHostConfig
		uhostConfigCopy.VPCId = info.vpcId
		uhostConfigCopy.SubnetId = info.subnetId
		uhostConfigCopy.PrivateIp = info.privateIp
		var uHostIds []string
		uHostIds, err = app.MigrateService.CreateUHost(&uhostConfigCopy, 1)
		if err != nil {
			return fmt.Errorf("[CreateUHost] about vpcId(%s):subnetId(%s):CubeId(%s):privateIp(%s) got error, %s",
				info.vpcId, info.subnetId, info.cubeId, info.privateIp, err)
		}
		if len(uHostIds) == 0 {
			return fmt.Errorf("[CreateUHost] about vpcId(%s):subnetId(%s):CubeId(%s):privateIp(%s), got no one uhost can be created",
				info.vpcId, info.subnetId, info.cubeId, info.privateIp)
		}
		// Waiting for UHost Running
		log.Logger.Sugar().Infof("[Waiting for UHost %s Running]", uHostIds[0])
		var runningUHostIds []string
		runningUHostIds, err = app.MigrateService.WaitingForUHostStatus(uHostIds, "Running")
		if err != nil {
			return fmt.Errorf("[Waiting for UHost %s Running] about vpcId(%s):subnetId(%s):CubeId(%s):privateIp(%s) got error, %s",
				uHostIds[0], info.vpcId, info.subnetId, info.cubeId, info.privateIp, err)
		}
		if len(runningUHostIds) == 0 {
			return fmt.Errorf("[Waiting for UHost %s Running] about vpcId(%s):subnetId(%s):CubeId(%s):privateIp(%s)， got no one uhost to be Running",
				uHostIds[0], info.vpcId, info.subnetId, info.cubeId, info.privateIp)
		}

		log.Logger.Sugar().Infof("[Migrate Cube %s to UHost %s Complete] with privateIp(%s)",
			info.cubeId, uHostIds[0], info.privateIp)
		successfulMigrateUHosts = append(successfulMigrateUHosts, uHostIds[0])
	}

	log.Logger.Sugar().Infof("[All Migrate Finished] Successful Create Running UHost ID List: %v", successfulMigrateUHosts)

	return nil
}

func (app *MigrateApp) tryCreateUHost(uhostConfig conf.UHostConfig, vpcId, subnetId string) error {
	uhostConfigCopy := uhostConfig
	uhostConfigCopy.VPCId = vpcId
	uhostConfigCopy.SubnetId = subnetId
	uhostConfigCopy.ChargeType = "Dynamic"

	uHostIds, err := app.MigrateService.CreateUHost(&uhostConfigCopy, 1)
	if err != nil {
		return fmt.Errorf("[Try to Create One UHost] got error, %s", err)
	}
	if len(uHostIds) == 0 {
		return fmt.Errorf("[Try to Create One UHost] got no one uhost can be created about the config, %v", uhostConfigCopy)
	}
	defer func(id string) {
		log.Logger.Sugar().Infof("[Cleanup the Try UHost %s]", id)
		if err = app.MigrateService.PowerOffUHostInstance(id); err != nil {
			log.Logger.Sugar().Warnf("[Cleanup the Try UHost %s] got error about poweroff uhost, %s", id, err)
		}
		_, err = app.MigrateService.WaitingForUHostStatus([]string{id}, "Stopped")
		if err != nil {
			log.Logger.Sugar().Warnf("[Cleanup the Try UHost %s] got error about waitting uhost stopped, %s", id, err)
		}
		if err = app.MigrateService.DeleteUHostInstance(id); err != nil {
			log.Logger.Sugar().Warnf("[Cleanup the Try UHost %s] got error about delete uhost, %s", id, err)
		}
		log.Logger.Sugar().Infof("[Cleanup the Try UHost %s Complete]", id)
	}(uHostIds[0])

	// Waiting for UHost Running
	log.Logger.Sugar().Infof("[Waiting for the Try UHost %s Running]", uHostIds[0])
	var runningUHostIds []string
	runningUHostIds, err = app.MigrateService.WaitingForUHostStatus(uHostIds, "Running")
	if err != nil {
		return fmt.Errorf("[Waiting for the Try UHost %s Running] got error, %s", uHostIds[0], err)
	}
	if len(runningUHostIds) == 0 {
		return fmt.Errorf("[Waiting for the Try UHost %s Running] "+
			"got no one uhost to be Running about the config, %v", uHostIds[0], uhostConfigCopy)
	}

	return nil
}
