package privateip

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ucloud/migrate/cmd/migrate/common"
	"github.com/ucloud/migrate/internal/app"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/migrate/internal/log"
	"github.com/ucloud/migrate/internal/utils"
	"os"
	"os/signal"
	"syscall"
)

func NewCmdPrivateIp() *cobra.Command {
	var cfgFile string
	var dryRun bool
	var config *conf.Config
	cmd := &cobra.Command{
		Use:   "private-ip",
		Short: "Make sure the private ip does not change during the migration process",
		PreRun: func(cmd *cobra.Command, args []string) {
			var err error
			config, err = common.InitConfig(cfgFile)
			utils.CheckErrorWithCode(err)
			if config.MigratePrivateIp == nil {
				utils.CheckErrorWithCode(fmt.Errorf("must set `migrate_private_ip` config for cmd `migrate private-ip`"))
			}

			if config.MigratePrivateIp.UHostConfig.VPCId != "" {
				utils.CheckErrorWithCode(fmt.Errorf("can not set " +
					"`migrate_ulb.uhost_config.vpc_id` config for cmd `migrate private-ip`, we will use the cube config"))
			}

			if config.MigratePrivateIp.UHostConfig.SubnetId != "" {
				utils.CheckErrorWithCode(fmt.Errorf("can not set " +
					"`migrate_ulb.uhost_config.subnet_id` config for cmd `migrate private-ip`, we will use the cube config"))
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			migrateApp, err := app.InitMigrateApp(*config)
			if err != nil {
				utils.CheckError(fmt.Errorf("init Migrate private Ip App got error, %s", err))
				return
			}

			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				for {
					<-c
					log.Logger.Sugar().Warnf("The Migrate private Ip will be down")
					os.Exit(1)
				}
			}()
			utils.CheckError(migrateApp.MigratePrivateIp(dryRun))
			return
		},
	}

	cmd.Flags().StringVarP(&cfgFile, "conf", "c", "", "config file like configs/config.json")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false,
		"dry run migrate, if set true which means that only try to create the uhost with config, waiting for uhost running and then delete it.")
	_ = cmd.MarkFlagRequired("conf")
	return cmd
}
