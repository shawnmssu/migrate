package eip

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

func NewCmdEIP() *cobra.Command {
	var cfgFile string
	var config *conf.Config
	cmd := &cobra.Command{
		Use:   "eip",
		Short: "Make sure the EIP(ExtranetIP) does not change during the migration process",
		PreRun: func(cmd *cobra.Command, args []string) {
			var err error
			config, err = common.InitConfig(cfgFile)
			utils.CheckErrorWithCode(err)
			if config.MigrateEIP == nil {
				utils.CheckErrorWithCode(fmt.Errorf("must set `migrate_eip` config for cmd `migrate eip`"))
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			migrateApp, err := app.InitMigrateApp(*config)
			if err != nil {
				utils.CheckError(fmt.Errorf("init Migrate EIP App got error, %s", err))
				return
			}

			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				for {
					<-c
					log.Logger.Sugar().Warnf("The Migrate EIP will be down")
					os.Exit(1)
				}
			}()
			utils.CheckError(migrateApp.MigrateEIP())
			return
		},
	}

	cmd.Flags().StringVarP(&cfgFile, "conf", "c", "", "config file like configs/config.json")
	_ = cmd.MarkFlagRequired("conf")
	return cmd
}
