package pkg

import (
	log "github.com/sirupsen/logrus"
	"pind/pkg/config"
)

const (
	ConfPathDef     = "/etc/pind/pind.yml"
	ConfPathBuiltIn = "built-in"
)

func loadConfigAndInit(ctx *Context) error {
	err := loadConfigFile(ctx)
	if err != nil {
		log.Errorf("loadConfig, loadConfigFile err = %v", err)
		return err
	}

	config0 := ctx.Config
	initLogger(config0.Log)

	str0, err := config.ToString0(config0)
	if err != nil {
		log.Errorf("loadConfig, config.ToString0 err = %v", err)
		return err
	}
	log.Infof("\n%s", str0)

	pool, err := NewPool(config0.Service.Pool)
	if err != nil {
		log.Errorf("loadConfig, NewPool err = %v", err)
		return err
	}
	ctx.pool = pool

	ctx.state = NewPinState()
	ctx.state.Idle = NewIdlePinCpu(ctx)

	return nil
}

func loadConfigFile(ctx *Context) error {
	confPath := ctx.ConfigPath

	if confPath == ConfPathBuiltIn {
		ctx.Config = config.NewDefaultConfig()
		log.Infof("use build-in config")
		return nil
	}

	config0, err := config.Load(confPath)
	if err != nil {
		log.Errorf("loadConfigFile, config.Load err = %v", err)
		return err
	}

	ctx.Config = config0
	return nil
}
