package common

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/ucloud/migrate/internal/conf"
	"github.com/ucloud/migrate/internal/log"
	"io/ioutil"
	"os"
)

func InitConfig(cfgFile string) (*conf.Config, error) {
	var config conf.Config
	if cfgFile == "" {
		return nil, fmt.Errorf("must set config by `--conf`")
	}

	confBytes, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("read config file got error, %v", err)
	}

	err = json.Unmarshal(confBytes, &config)
	if err != nil {

		return nil, fmt.Errorf("unmarshal config got error, %v", err)
	}

	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		return nil, fmt.Errorf("config validator got error, %v", err)
	}

	if config.PublicKey == "" {
		if os.Getenv("UCLOUD_PUBLIC_KEY") != "" {
			config.PublicKey = os.Getenv("UCLOUD_PUBLIC_KEY")
		} else {
			return nil, fmt.Errorf("must set `public_key` by config or env")
		}
	}

	if config.PrivateKey == "" {
		if os.Getenv("UCLOUD_PRIVATE_KEY") != "" {
			config.PrivateKey = os.Getenv("UCLOUD_PRIVATE_KEY")
		} else {
			return nil, fmt.Errorf("must set `private_key` by config or env")
		}
	}

	if config.ProjectId == "" {
		if os.Getenv("UCLOUD_PROJECT_ID") != "" {
			config.ProjectId = os.Getenv("UCLOUD_PROJECT_ID")
		} else {
			return nil, fmt.Errorf("must set `project_id` by config or env")
		}
	}

	if config.Region == "" {
		if os.Getenv("UCLOUD_REGION") != "" {
			config.Region = os.Getenv("UCLOUD_REGION")
		} else {
			return nil, fmt.Errorf("must set `region` by config or env")
		}
	}

	if err = log.InitLogger(config.Log); err != nil {
		return nil, fmt.Errorf("init logger got error, %s", err)
	}

	return &config, nil
}
