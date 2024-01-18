package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"pind/pkg/config"
	"pind/pkg/utils/os_utils"
)

const (
	DefConfPath = "/etc/pind/pind.conf"
)

func loadConfigAndInit(ctx *Context) error {
	err := loadConfigFile(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "loadConfig, loadConfigFile = err = %v", err)
		return err
	}

	config0 := ctx.Config
	initLogger(config0.Log)

	return nil
}

func loadConfigFile(ctx *Context) error {
	confPath := ctx.ConfigPath

	same, err := os_utils.SameFiles(confPath, DefConfPath)
	if err == nil && same {
		ctx.Config = config.NewDefaultConfig()
		log.Debugf("use build-in config")
		return nil
	}

	config0, err := config.Load(confPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "loadConfigFile, config.Load = err = %v", err)
		return err
	}

	ctx.Config = config0
	return nil
}
