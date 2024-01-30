package pkg

import (
	log "github.com/sirupsen/logrus"
	"pind/pkg/config"
	"pind/pkg/http_api"
	"pind/pkg/numa"
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

	err = checkConfig(config0)
	if err != nil {
		log.Errorf("loadConfig, checkConfig err = %v", err)
		return err
	}

	pool, err := NewPool(config0.Service.Pool)
	if err != nil {
		log.Errorf("loadConfig, NewPool err = %v", err)
		return err
	}
	ctx.pool = pool

	ctx.state = NewPinState()
	ctx.state.Idle = NewIdlePinCpu(ctx)

	ctx.HttpApi = http_api.NewHttpApi(config0.Service.HttpApi)

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

func checkConfig(config0 *config.Config) error {
	idle := config0.Service.Pool.Idle.Values
	err := numa.IsCpusOnSameNumaNode(idle)
	if err != nil {
		log.Errorf("checkConfig, pool idle numa.IsCpusOnSameNumaNode err = %v", err)
		return err
	}

	return nil
}
