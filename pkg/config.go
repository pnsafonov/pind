package pkg

import (
	"fmt"
	"github.com/pnsafonov/pind/pkg/config"
	"github.com/pnsafonov/pind/pkg/http_api"
	"github.com/pnsafonov/pind/pkg/numa"
	"github.com/pnsafonov/pind/pkg/utils/os_utils"
	log "github.com/sirupsen/logrus"
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
	initLogger(config0.Log, ctx.Service)

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
	config0, err := loadConfigFile0(ctx)
	if err != nil {
		log.Errorf("loadConfigFile, config.Load err = %v", err)
		return err
	}

	err = initDefaultConfig(config0)
	if err != nil {
		log.Errorf("loadConfigFile, initDefaultConfig err = %v", err)
		return err
	}

	ctx.Config = config0
	return nil
}

func loadConfigFile0(ctx *Context) (*config.Config, error) {
	confPath := ctx.ConfigPath

	if confPath == ConfPathBuiltIn {
		config0 := config.NewDefaultConfig(ctx.Service)
		log.Infof("use build-in config")
		return config0, nil
	}

	config0, err := config.Load(confPath, ctx.Service)
	if err != nil {
		log.Errorf("loadConfigFile, config.Load err = %v", err)
		return nil, err
	}

	return config0, nil
}

func initDefaultConfig(config0 *config.Config) error {
	err := initDefaultNumaConfig(config0)
	if err != nil {
		log.Errorf("initDefaultConfig, initDefaultNumaConfig err = %v", err)
		return err
	}

	err = initDefaultFiltersConfig(config0)
	if err != nil {
		log.Errorf("initDefaultConfig, initDefaultFiltersConfig err = %v", err)
		return err
	}

	if config0.Service.Pool.LoadType == config.Phys {
		// convert idle phys cores to logical
		// then load_type: "phys"
		logicalCores := numa.PhysCoresToLogical(config0.Service.Pool.Idle.Values)
		config0.Service.Pool.Idle.Values = logicalCores
	}

	return nil
}

func initDefaultNumaConfig(config0 *config.Config) error {
	if len(config0.Service.Pool.Idle.Values) != 0 && len(config0.Service.Pool.Load.Values) != 0 {
		return nil
	}

	nodesPhys, err := numa.GetNodesPhys()
	if err != nil {
		return err
	}
	l0 := len(nodesPhys[0].Cores[0].ThreadSiblings)
	if l0 == 0 {
		// no thread siblings, no hyper-threading
		return initDefaultLogicalPoolConfig(config0)
	}

	return initDefaultPhysPoolConfig(config0)
}

func initDefaultPhysPoolConfig(config0 *config.Config) error {
	nodesPhys, err := numa.GetNodesPhys()
	if err != nil {
		return err
	}

	node0 := nodesPhys[0]

	numaNodesCount := len(nodesPhys)
	numaCoresCount := len(node0.Cores)
	coresCount := getIdleCoresCountDefault(numaNodesCount, numaCoresCount)

	idleCores := make([]int, 0, coresCount)
	for i := 0; i < coresCount; i++ {
		id := node0.Cores[i].Id
		idleCores = append(idleCores, id)
	}

	l1 := (numaNodesCount-1)*numaCoresCount + numaCoresCount - coresCount
	loadCores := make([]int, 0, l1)

	for i := coresCount; i < numaCoresCount; i++ {
		id := node0.Cores[i].Id
		loadCores = append(loadCores, id)
	}

	for i := 1; i < numaNodesCount; i++ {
		node := nodesPhys[i]
		for _, core := range node.Cores {
			loadCores = append(loadCores, core.Id)
		}
	}

	pool := config0.Service.Pool
	pool.Idle.Values = idleCores
	pool.Load.Values = loadCores
	pool.LoadType = config.Phys

	config0.Service.Pool = pool
	return nil
}

func initDefaultLogicalPoolConfig(config0 *config.Config) error {
	nodes, err := numa.GetNodes()
	if err != nil {
		return err
	}

	node0 := nodes[0]

	numaNodesCount := len(nodes)
	numaCoresCount := len(node0.Cpus)
	coresCount := getIdleCoresCountDefault(numaNodesCount, numaCoresCount)

	idleCores := make([]int, 0, coresCount)
	for i := 0; i < coresCount; i++ {
		id := node0.Cpus[i]
		idleCores = append(idleCores, id)
	}

	l1 := (numaNodesCount-1)*numaCoresCount + numaCoresCount - coresCount
	loadCores := make([]int, 0, l1)

	for i := coresCount; i < numaCoresCount; i++ {
		id := node0.Cpus[i]
		loadCores = append(loadCores, id)
	}

	for i := 1; i < numaNodesCount; i++ {
		node := nodes[i]
		for _, core := range node.Cpus {
			loadCores = append(loadCores, core)
		}
	}

	pool := config0.Service.Pool
	pool.Idle.Values = idleCores
	pool.Load.Values = loadCores
	pool.LoadType = config.Logical

	config0.Service.Pool = pool
	return nil
}

func initDefaultFiltersConfig(config0 *config.Config) error {
	l0 := len(config0.Service.Filters0)
	l1 := len(config0.Service.Filters1)
	l2 := len(config0.Service.FiltersAlwaysIdle)
	if l0 != 0 && l1 != 0 && l2 != 0 {
		return nil
	}

	paths0, ok := os_utils.Which0("kvm", "qemu-system-x86_64")
	if !ok {
		return fmt.Errorf("kvm, qemu-system-x86_64 not found in $PATH")
	}
	filters0 := config.NewDefaultFilters2(paths0)
	filters1 := config.NewDefaultFilters2(paths0)

	paths1 := []string{"/usr/local/bin/node_exporter"}
	pathsAlways := config.NewDefaultFilters2(paths1)

	config0.Service.Filters0 = filters0
	config0.Service.Filters1 = filters1
	config0.Service.FiltersAlwaysIdle = pathsAlways
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
