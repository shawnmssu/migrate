package service

import (
	"github.com/ucloud/migrate/internal/client"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/ucloud-sdk-go/services/cube"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
)

type MigrateCubeService interface {
	UHostService
	Cube
}

type UHostService interface {
	BindEIPToUHost(uhostId, eipId string) error
	CreateUHost(config *conf.UHostConfig, maxCount int) ([]string, error)
	DescribeUHostById(uhostId string) (*uhost.UHostInstanceSet, error)
}

type Cube interface {
	UnBindEIPToCube(cubeId, eipId string) error
	BindEIPToCube(cubeId, eipId string) error
	GetCubeCubePodExtendInfoList(config *conf.CubeConfig) ([]cube.CubeExtendInfo, error)
}

func NewBasicMigrateCubeService(client *client.UCloudClient) MigrateCubeService {
	return client
}
