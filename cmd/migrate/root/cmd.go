package root

import (
	"github.com/spf13/cobra"
	"github.com/ucloud/migrate/cmd/migrate/eip"
	"github.com/ucloud/migrate/cmd/migrate/privateip"
	"github.com/ucloud/migrate/cmd/migrate/ulb"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "migrate Cube to UHost",
	}

	cmd.AddCommand(
		ulb.NewCmdULB(),
		eip.NewCmdEIP(),
		privateip.NewCmdPrivateIp(),
	)
	return cmd
}
