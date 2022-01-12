package service

import (
	"github.com/ucloud/migrate/internal/client"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/ucloud-sdk-go/services/cube"
)

type MigrateCubeService interface {
	UHostService
	Cube
}

type UHostService interface {
	BindEIPToUHost(uHostId, eipId string) error
	UnBindEIPToUHost(uHostId, eipId string) error
	CreateUHost(config *conf.UHostConfig, count int) ([]string, error)
	WaitingForUHostRunning(uHostIds []string) ([]string, error)
}

type Cube interface {
	UnBindEIPToCube(cubeId, eipId string) error
	BindEIPToCube(cubeId, eipId string) error
	GetCubeCubePodExtendInfoList(config *conf.CubeConfig) ([]cube.CubeExtendInfo, error)
}

func NewBasicMigrateCubeService(client *client.UCloudClient) MigrateCubeService {
	return client
}
