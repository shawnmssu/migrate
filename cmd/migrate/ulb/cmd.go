package ulb

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

func NewCmdULB() *cobra.Command {
	var cfgFile string
	var config *conf.Config
	cmd := &cobra.Command{
		Use:   "ulb",
		Short: "Make sure the ULB(LoadBalance) does not change during the migration process",
		PreRun: func(cmd *cobra.Command, args []string) {
			var err error
			config, err = common.InitConfig(cfgFile)
			utils.CheckErrorWithCode(err)
			if config.MigrateULB == nil {
				utils.CheckErrorWithCode(fmt.Errorf("must set `migrate_ulb` config for cmd `migrate ulb`"))
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			migrateApp, err := app.InitMigrateApp(*config)
			if err != nil {
				utils.CheckError(fmt.Errorf("init Migrate ULB got error, %s", err))
				return
			}

			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				for {
					<-c
					log.Logger.Sugar().Warnf("The Migrate ULB will be down")
					os.Exit(1)
				}
			}()
			utils.CheckError(migrateApp.MigrateULB())
		},
	}
	cmd.Flags().StringVarP(&cfgFile, "conf", "c", "", "config file like configs/config.json")
	_ = cmd.MarkFlagRequired("conf")
	return cmd
}
