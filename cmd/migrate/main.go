package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/ucloud/migrate/internal/app"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/migrate/internal/log"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/config.json", "config path, eg: -conf config.json")
	flag.Parse()
}

func main() {
	flag.Parse()

	confBytes, err := ioutil.ReadFile(flagConf)
	if err != nil {
		fmt.Printf("read config file got error, %v\n", err)
		return
	}
	var bc conf.Config
	err = json.Unmarshal(confBytes, &bc)
	if err != nil {
		fmt.Printf("unmarshal config got error, %v\n", err)
		return
	}

	validate := validator.New()
	err = validate.Struct(bc)
	if err != nil {
		fmt.Printf("config validator got error, %v\n", err)
		return
	}

	if bc.PublicKey == "" {
		if os.Getenv("UCLOUD_PUBLIC_KEY") != "" {
			bc.PublicKey = os.Getenv("UCLOUD_PUBLIC_KEY")
		} else {
			fmt.Println("must set `public_key` by config or env")
			return
		}
	}

	if bc.PrivateKey == "" {
		if os.Getenv("UCLOUD_PRIVATE_KEY") != "" {
			bc.PrivateKey = os.Getenv("UCLOUD_PRIVATE_KEY")
		} else {
			fmt.Println("must set `private_key` by config or env")
			return
		}
	}

	if bc.ProjectId == "" {
		if os.Getenv("UCLOUD_PROJECT_ID") != "" {
			bc.ProjectId = os.Getenv("UCLOUD_PROJECT_ID")
		} else {
			fmt.Println("must set `project_id` by config or env")
			return
		}
	}

	if bc.Region == "" {
		if os.Getenv("UCLOUD_REGION") != "" {
			bc.Region = os.Getenv("UCLOUD_REGION")
		} else {
			fmt.Println("must set `region` by config or env")
			return
		}
	}

	if err = log.InitLogger(bc.Log); err != nil {
		fmt.Printf("init logger got error, %s\n", err)
		return
	}

	migrateApp := app.NewMigrateApp(bc)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			<-c
			log.Logger.Sugar().Warnf("The migrate will be down")
			os.Exit(1)
		}
	}()

	migrateApp.Run()
}
