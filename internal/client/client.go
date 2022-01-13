package client

import (
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/ucloud-sdk-go/services/cube"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
)

type UCloudClient struct {
	UHostConn *uhost.UHostClient
	CubeConn  *cube.CubeClient
	UNetConn  *unet.UNetClient
	ULBConn   *ulb.ULBClient
}

func NewClient(sdkCfg conf.Config) (*UCloudClient, error) {
	cfg := ucloud.NewConfig()
	cfg.Region = sdkCfg.Region
	cfg.ProjectId = sdkCfg.ProjectId
	cfg.LogLevel = log.PanicLevel
	cred := auth.NewCredential()
	cred.PublicKey = sdkCfg.PublicKey
	cred.PrivateKey = sdkCfg.PrivateKey
	var client UCloudClient
	client.UHostConn = uhost.NewClient(&cfg, &cred)
	client.CubeConn = cube.NewClient(&cfg, &cred)
	client.UNetConn = unet.NewClient(&cfg, &cred)
	client.ULBConn = ulb.NewClient(&cfg, &cred)

	return &client, nil
}
