package pkg

import (
	"fmt"
	"github.com/pnsafonov/pind/pkg/config"
	"github.com/pnsafonov/pind/pkg/http_api"
	"github.com/pnsafonov/pind/pkg/numa"
	"github.com/prometheus/procfs"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const userHZ = 100

type Context struct {
	ConfigPath  string
	Service     bool
	PrintConfig bool

	Config *config.Config

	Done chan int

	pool            *Pool
	lastAll         []*ProcInfo
	lastInFilter    []*ProcInfo
	lastNotInFilter []*ProcInfo
	alwaysIdle      []*ProcInfo

	state PinState

	lastCpuInfo *numa.Info

	HttpApi *http_api.HttpApi

	Version string
	GitHash string

	firstRun bool // first work cycle
}

func NewContext(version string, gitHash string, isService bool) *Context {
	ctx := &Context{}
	ctx.Done = make(chan int, 1)
	ctx.ConfigPath = ConfPathDef
	ctx.Service = isService
	ctx.PrintConfig = false
	ctx.Config = config.NewDefaultConfig(isService)
	ctx.Version = version
	ctx.GitHash = gitHash
	ctx.firstRun = true
	return ctx
}

func logContext(ctx *Context) {
	log.Infof("ConfigPath  = %s", ctx.ConfigPath)
	log.Infof("Service     = %v", ctx.Service)
	log.Infof("PrintConfig = %v", ctx.PrintConfig)
}

func RunService(ctx *Context) error {
	logContext(ctx)

	err := loadConfigAndInit(ctx)
	if err != nil {
		log.Errorf("RunService, loadConfigAndInit err = %v", err)
		return err
	}

	err = pinSelfProcess(ctx)
	if err != nil {
		log.Errorf("RunService, pinSelfProcess err = %v", err)
		return err
	}

	err = doHttpApi(ctx)
	if err != nil {
		log.Errorf("RunService, doHttpApi err = %v", err)
		return err
	}

	doSignals(ctx)
	doLoop(ctx)

	return nil
}

// pinSelfProcess - привязать все потоки текущего процесса
// к idle ядрам
func pinSelfProcess(ctx *Context) error {
	pid := os.Getpid()
	log.Printf("pid = %d", pid)
	set := &ctx.state.Idle.CpuSet

	threads, err := procfs.AllThreads(pid)
	if err != nil {
		log.Errorf("pinSelfProcess, procfs.AllThreads err = %v, pid = %d", err, pid)
		return err
	}
	l1 := len(threads)
	for j := 0; j < l1; j++ {
		thread := threads[j]
		err := unix.SchedSetaffinity(thread.PID, set)
		if err != nil {
			log.Errorf("pinSelfProcess, schedSetaffinity err = %v, thread pid = %d", err, thread.PID)
			return err
		}
	}

	return nil
}

func doHttpApi(ctx *Context) error {
	if !ctx.HttpApi.Config.Enabled {
		return nil
	}

	err := ctx.HttpApi.GoServe()
	if err != nil {
		log.Errorf("doHttpApi, ctx.HttpApi.GoServe err = %v", err)
		return err
	}

	return nil
}

func doSignals(ctx *Context) {
	done := ctx.Done

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go signalsLoop(sigChan, done)
}

func signalsLoop(sigChan chan os.Signal, done chan int) {
	for {
		sig := <-sigChan
		log.Infof("signalsLoop, received signal = %d", sig)
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			done <- 1
			return
		case syscall.SIGHUP:
			// reload config
			log.Infof("signalsLoop, SIGHUP received, reload conf (not implemented yet)")
		default:
			return
		}
	}
}

func doLoop(ctx *Context) {
	interval := time.Duration(ctx.Config.Service.Interval)
	done := ctx.Done

	log.Infof("doLoop, pind started as daemon, version = %s, git_hash = %s", ctx.Version, ctx.GitHash)
	ticker := time.NewTicker(interval * time.Millisecond)
	for {
		select {
		case <-done:
			log.Infof("doLoop, received done, stopping pind, version = %s, git_hash = %s", ctx.Version, ctx.GitHash)
			return
		case t := <-ticker.C:
			handler(ctx, t)
		}
	}
}

func setTime(procs []*ProcInfo, time0 time.Time) {
	l0 := len(procs)
	for i := 0; i < l0; i++ {
		proc := procs[i]
		proc.time = time0
	}
}

func calcTimeDelta(prev []*ProcInfo, cur []*ProcInfo) float64 {
	timePrev := prev[0].time
	timeCur := cur[0].time
	delta := timeCur.Sub(timePrev)
	return delta.Seconds()
}

func printProcs0(procs []*ProcInfo, time0 time.Time) {
	l0 := len(procs)
	for i := 0; i < l0; i++ {
		proc := procs[i]
		fmt.Printf("% 6d %s %d\n", proc.Proc.PID, proc.Stat.Comm, proc.Stat.Starttime)
	}
}

func printProcs1(procs []*ProcInfo, time0 time.Time, timeDelta float64) {
	l0 := len(procs)
	for i := 0; i < l0; i++ {
		proc := procs[i]
		fmt.Printf("% 6d %s %d cpu=%.0f delta=%.0f\n", proc.Proc.PID, proc.Stat.Comm, proc.Stat.Starttime, proc.cpu0, timeDelta)
	}
}

func handler(ctx *Context, time0 time.Time) {
	errs := ctx.state.Errors
	errs.clear()

	err0 := calcProcsCPU(ctx, time0)
	if err0 != nil {
		log.Errorf("handler, calcProcsCPU err = %v", err0)
		errs.CalcProcsCPU = err0
	}

	err5 := calcCoresCPU(ctx)
	if err5 != nil {
		log.Errorf("handler, calcCoresCPU err = %v", err5)
		errs.CalcCoresCPU = err5
	}

	ctx.state.UpdateProcs(ctx.lastInFilter)

	err6 := pinAlwaysIdle(ctx)
	if err6 != nil {
		log.Errorf("handler, pinAlwaysIdle err = %v", err6)
	}

	err1 := pinNotInFilterToIdle(ctx)
	if err1 != nil {
		log.Errorf("handler, pinNotInFilterToIdle err = %v", err1)
		errs.PinNotInFilterToIdle = err1
	}

	if ctx.Config.Service.Pool.PinMode == config.PinModeNormal || ctx.firstRun {
		err2 := ctx.state.PinIdle()
		if ctx.firstRun {
			ctx.firstRun = false
		}
		if err2 != nil {
			log.Errorf("handler, ctx.state.PinLoad err = %v", err2)
			errs.StatePinIdle = err2
		}
	}

	err3 := ctx.state.PinLoad(ctx)
	if err3 != nil {
		log.Errorf("handler, ctx.state.PinLoad err = %v", err3)
		errs.StatePinLoad = err3
	}

	err4 := setHttpApiData(ctx)
	if err4 != nil {
		log.Errorf("handler, setHttpApiData err = %v", err4)
	}
}

func calcProcsCPU(ctx *Context, time0 time.Time) error {
	filters0 := ctx.Config.Service.Filters0
	filters1 := ctx.Config.Service.Filters1
	filtersAlwaysIdle := ctx.Config.Service.FiltersAlwaysIdle
	threshold := ctx.Config.Service.Threshold
	ignore := ctx.Config.Service.Ignore

	procsAll, procsAlwaysIdle, err := filterProcsInfo0(filters0, filtersAlwaysIdle, ignore)
	if err != nil {
		log.Errorf("calcProcsCPU, filterProcsInfo0 err = %v", err)
		return err
	}
	setTime(procsAll, time0)
	//printProcs0(procsAll, time0)

	prevLastAll := ctx.lastAll
	l0 := len(prevLastAll)
	if l0 == 0 {
		// first run, nothing to do
		inFilter, notInFilter := filterProcInfo(procsAll, filters1)
		ctx.lastAll = procsAll
		ctx.lastInFilter = inFilter
		ctx.lastNotInFilter = notInFilter
		return nil
	}

	l1 := len(procsAll)
	if l1 == 0 {
		ctx.lastAll = nil
		ctx.lastInFilter = nil
		ctx.lastNotInFilter = nil
		return nil
	}

	timeDelta := calcTimeDelta(prevLastAll, procsAll)

	// second filtration
	inFilter, notInFilter := filterProcInfo(procsAll, filters1)
	l2 := len(inFilter)
	for i := 0; i < l2; i++ {
		proc1 := inFilter[i]
		// proc0
		proc0, ok := getSameProc(prevLastAll, proc1)
		if !ok {
			continue
		}

		cpu0 := calcProcCpuLoad0(proc0, proc1, timeDelta)
		proc1.cpu0 = cpu0
		proc1.load = cpu0 > threshold
	}
	ctx.lastAll = procsAll
	ctx.lastInFilter = inFilter
	ctx.lastNotInFilter = notInFilter
	//printProcs1(procsAll, time0, timeDelta)
	ctx.alwaysIdle = procsAlwaysIdle

	return nil
}

func calcCoresCPU(ctx *Context) error {
	cpuInfos, err := numa.GetCpuInfos()
	if err != nil {
		log.Errorf("calcCoresCPU, numa.GetCpuInfos err = %v", err)
		return err
	}

	prevCpuInfos := ctx.lastCpuInfo
	if prevCpuInfos == nil {
		ctx.lastCpuInfo = cpuInfos
		return nil
	}

	l0 := len(prevCpuInfos.Nodes)
	l1 := len(cpuInfos.Nodes)
	if l0 != l1 {
		log.Warningf("calcCoresCPU, different numa nodes count prev = %d, cur = %d", l0, l1)
	}

	for i := 0; i < l0 && i < l1; i++ {
		node := cpuInfos.Infos[i]
		prevNode := prevCpuInfos.Infos[i]

		for cpu, cpuInfo := range node.Cores {
			prevCpuInfo, ok := prevNode.Cores[cpu]
			if !ok {
				log.Warningf("calcCoresCPU, can't find prevCpuInfo with cpu = %d", cpu)
				continue
			}

			cpuLoad := calcCoreCpuLoad0(prevCpuInfo, cpuInfo)
			cpuInfo.CpuLoad = cpuLoad
		}
	}

	calcIdlePoolLoad(ctx, cpuInfos)
	calcPoolLoad(ctx, cpuInfos)

	ctx.lastCpuInfo = cpuInfos
	return nil
}

func calcProcCpuLoad0(prev *ProcInfo, cur *ProcInfo, timeDelta float64) float64 {
	if prev == nil || cur == nil || timeDelta <= 0 {
		return 0
	}

	totalS := cur.Stat.STime - prev.Stat.STime
	totalU := cur.Stat.UTime - prev.Stat.UTime
	total0 := float64(totalS+totalU) / userHZ

	cpu0 := total0 / timeDelta
	cpu1 := cpu0 * 100

	return cpu1
}

func calcCoreCpuLoad0(prev *numa.CpuInfo, cur *numa.CpuInfo) float64 {
	if prev == nil || cur == nil {
		return 0
	}

	curSum := cur.GetSumm()
	prevSum := prev.GetSumm()
	delta := curSum - prevSum

	curSum0 := cur.GetSumm0()
	prevSum0 := prev.GetSumm0()
	delta0 := curSum0 - prevSum0

	load := delta0 / delta
	load *= 100

	return load
}

func calcIdlePoolLoad(ctx *Context, cpuInfo *numa.Info) {
	pool := ctx.pool
	idle := pool.Config.Idle.Values
	idleOverwork := ctx.Config.Service.IdleOverwork

	load := float64(0)
	for _, cpu := range idle {
		info, ok := cpuInfo.GetCpuInfo(cpu)
		if !ok {
			log.Warningf("calcIdlePoolLoad, cpuInfo.GetCpuInfo failed for cpu = %d", cpu)
			continue
		}
		load += info.CpuLoad
	}

	pool.IdleLoad0 = load

	load1 := float64(0)
	if pool.IdleLoadFull0 != 0 {
		load1 = pool.IdleLoad0 / pool.IdleLoadFull0
	}
	load1 *= 100
	pool.IdleLoad1 = load1

	if load1 >= idleOverwork {
		log.Warningf("calcIdlePoolLoad, idle_overwork is high %.2f >= %.2f %%", load1, idleOverwork)
		ctx.state.Errors.IdleOverwork = fmt.Errorf("idle_overwork %.2f is greater than %.2f %%", load1, idleOverwork)
	}
}

func calcPoolLoad(ctx *Context, cpuInfo *numa.Info) {
	pool := ctx.pool

	pool.LoadFree0 = 0
	pool.LoadFree1 = 0
	pool.LoadFreeFull = 0
	pool.LoadUsed0 = 0
	pool.LoadUsed1 = 0
	pool.LoadUsedFull = 0

	for _, node := range pool.Nodes {

		//load := float64(0)
		//for _, cpu := range node.Node.Cpus {
		//	info, ok := cpuInfo.GetCpuInfo(cpu)
		//	if !ok {
		//		log.Warningf("calcPoolLoad, cpuInfo.GetCpuInfo failed for cpu = %d", cpu)
		//		continue
		//	}
		//	load += info.CpuLoad
		//}
		//node.Load0 = load

		loadFree0, loadFree1, loadFreeFull := calcPoolCoresLoad(node.LoadFree, cpuInfo)
		loadUsed0, loadUsed1, loadUsedFull := calcPoolCoresLoad(node.LoadUsed, cpuInfo)

		node.LoadFree0 = loadFree0
		node.LoadFree1 = loadFree1
		node.LoadFreeFull = loadFreeFull

		node.LoadUsed0 = loadUsed0
		node.LoadUsed1 = loadUsed1
		node.LoadUsedFull = loadUsedFull

		pool.LoadFree0 += loadFree0
		pool.LoadFreeFull += loadFreeFull
		pool.LoadUsed0 += loadUsed0
		pool.LoadUsedFull += loadUsedFull
	}

	if pool.LoadFreeFull != 0 {
		pool.LoadFree1 = (pool.LoadFree0 / pool.LoadFreeFull) * 100
	}
	if pool.LoadUsedFull != 0 {
		pool.LoadUsed1 = (pool.LoadUsed0 / pool.LoadUsedFull) * 100
	}
}

func calcPoolCoresLoad(poolCores map[int]*PoolCore, cpuInfo *numa.Info) (float64, float64, float64) {
	l0 := len(poolCores)
	loadFull := float64(l0 * 100)

	load := float64(0)
	for _, poolCore := range poolCores {
		info, ok := cpuInfo.GetCpuInfo(poolCore.Id)
		if !ok {
			log.Warningf("calcPoolCoresLoad, cpuInfo.GetCpuInfo failed for cpu = %d", poolCore.Id)
			continue
		}

		load += info.CpuLoad
	}

	load1 := float64(0)
	if l0 != 0 {
		load1 = (load / loadFull) * 100
	}

	return load, load1, loadFull
}
