package app

import (
	"fmt"
	"github.com/ucloud/migrate/internal/client"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/migrate/internal/service"
)

type MigrateApp struct {
	Config         conf.Config
	MigrateService service.MigrateCubeService
}

func InitMigrateApp(bc conf.Config) (*MigrateApp, error) {
	globalClient, err := client.NewClient(bc)
	if err != nil {
		return nil, fmt.Errorf("new SDK Clinet Got error, %s", err)
	}

	return &MigrateApp{
		bc,
		service.NewBasicMigrateCubeService(globalClient),
	}, nil
}
