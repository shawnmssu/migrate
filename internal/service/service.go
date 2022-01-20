package service

import (
	"github.com/ucloud/migrate/internal/client"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/ucloud-sdk-go/services/cube"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
)

type MigrateCubeService interface {
	UHostService
	CubeService
	ULBService
}

type UHostService interface {
	BindEIPToUHost(uHostId, eipId string) error
	UnBindEIPToUHost(uHostId, eipId string) error
	CreateUHost(config *conf.UHostConfig, count int) ([]string, error)
	WaitingForUHostStatus(uHostIds []string, status string) ([]string, error)
	PowerOffUHostInstance(uhostId string) error
	DeleteUHostInstance(uhostId string) error
	DescribeUHostById(uhostId string) (*uhost.UHostInstanceSet, error)
}

type CubeService interface {
	UnBindEIPToCube(cubeId, eipId string) error
	BindEIPToCube(cubeId, eipId string) error
	GetCubePodExtendInfoList(config *conf.CubeConfig) ([]cube.CubeExtendInfo, error)
	GetCubePodInfoList(config *conf.CubeConfig) ([]client.CubePodInfo, error)
	DeleteCube(cubeId string) error
}

type ULBService interface {
	UnBindBackendToUlB(ulbId, backendId string) error
	BindUHostToUlBVServer(cubeId, vServerId, uHostId string, port int) (string, error)
	GetULBVServerInfoListAboutCube(ulbId string) ([]client.VServerCubeInfo, error)
	DescribeBackendById(lbId, vServerId, backendId string) (*ulb.ULBBackendSet, error)
}

func NewBasicMigrateCubeService(client *client.UCloudClient) MigrateCubeService {
	return client
}
