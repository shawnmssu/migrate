package main

import (
	"github.com/ucloud/migrate/cmd/migrate/root"
	"github.com/ucloud/migrate/internal/utils"
)

func main() {
	utils.CheckError(root.NewCmd().Execute())
}
