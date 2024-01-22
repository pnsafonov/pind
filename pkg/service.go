package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
	"pind/pkg/config"
	"syscall"
	"time"
)

const userHZ = 100

type Context struct {
	ConfigPath string
	Service    bool

	Config *config.Config

	Done chan int

	pool *Pool
	last []*ProcInfo

	state PinState
}

func NewContext() *Context {
	ctx := &Context{}
	ctx.Done = make(chan int, 1)
	ctx.ConfigPath = DefConfPath
	ctx.Service = false
	ctx.Config = config.NewDefaultConfig()
	return ctx
}

func RunService(ctx *Context) error {
	err := loadConfigAndInit(ctx)
	if err != nil {
		log.Errorf("RunService, loadConfigAndInit err = %v", err)
		return err
	}

	doSignals(ctx)
	doLoop(ctx)

	return nil
}

func doSignals(ctx *Context) {
	done := ctx.Done

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go signalsLoop(sigChan, done)
}

func signalsLoop(sigChan chan os.Signal, done chan int) {
	for {
		sig := <-sigChan
		log.Infof("received signal = %d", sig)
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			done <- 1
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func doLoop(ctx *Context) {
	interval := time.Duration(ctx.Config.Service.Interval)
	done := ctx.Done

	ticker := time.NewTicker(interval * time.Millisecond)
	for {
		select {
		case <-done:
			log.Infof("doLoop received done")
			return
		case t := <-ticker.C:
			//log.Infof("Before sleep %v", t)
			//time.Sleep(3 * time.Second)
			//log.Infof("After Sleep  %v", t)

			_ = handler(ctx, t)
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

func handler(ctx *Context, time0 time.Time) error {
	err := calcCPU(ctx, time0)
	if err != nil {
		log.Errorf("handler, calcCPU err = %v", err)
		return err
	}

	//pinProcs(ctx, ctx.last)
	ctx.state.UpdateProcs(ctx.last)
	err = ctx.state.PinCores(ctx)
	if err != nil {
		log.Errorf("handler, ctx.state.PinCores err = %v", err)
		return err
	}

	return nil
}

func calcCPU(ctx *Context, time0 time.Time) error {
	filters := ctx.Config.Service.Filters

	procs1, err := filterProcsInfo0(filters)
	if err != nil {
		log.Errorf("calcCPU, filterProcsInfo0 err = %v", err)
		return err
	}
	setTime(procs1, time0)
	//printProcs0(procs1, time0)

	procs0 := ctx.last
	l0 := len(ctx.last)
	if l0 == 0 {
		// first run, nothing to do
		ctx.last = procs1
		return nil
	}

	l1 := len(procs1)
	if l1 == 0 {
		return nil
	}

	timeDelta := calcTimeDelta(procs0, procs1)

	for i := 0; i < l1; i++ {
		proc1 := procs1[i]
		// proc0
		proc0, ok := getSameProc(procs0, proc1)
		if !ok {
			continue
		}

		cpu0 := calcCpuLoad0(proc0, proc1, timeDelta)
		proc1.cpu0 = cpu0
	}
	ctx.last = procs1
	//printProcs1(procs1, time0, timeDelta)

	return nil
}

func calcCpuLoad0(prev *ProcInfo, cur *ProcInfo, timeDelta float64) float64 {
	if prev == nil || cur == nil || timeDelta <= 0 {
		return 0
	}

	totalS := cur.Stat.STime - prev.Stat.STime
	totalU := cur.Stat.UTime - prev.Stat.UTime
	total0 := float64(totalS+totalU) / userHZ

	cpu0 := total0 / timeDelta
	cpu1 := cpu0 * 100
	if cpu1 > 100 {
		cpu1 = 100
	}
	return cpu1
}

func pinProcs(ctx *Context, procs []*ProcInfo) {
	threshold := ctx.Config.Service.Threshold

	used := make(map[int]byte)
	l0 := len(procs)
	for i := 0; i < l0; i++ {
		proc := procs[i]

		isLoad := proc.cpu0 >= threshold
		if !isLoad {
			// idle
			_ = markOnIdle(ctx, proc)
			continue
		}

		markOnLoad(ctx, proc, used)
	}

	for i := 0; i < l0; i++ {
		proc := procs[i]

		isLoad := proc.cpu0 >= threshold
		if !isLoad {
			// idle
			pinOnIdle(ctx, proc)
			continue
		}

		pinOnLoad(ctx, proc)
	}
}

func markOnIdle(ctx *Context, proc *ProcInfo) error {
	var err error
	l0 := len(proc.Threads)
	for i := 0; i < l0; i++ {
		thread := proc.Threads[i]

		if isMasksEqual(thread.CpuSet, ctx.pool.IdleMask) {
			continue
		}

		err0 := unix.SchedSetaffinity(thread.Thread.PID, &ctx.pool.IdleMask)
		if err0 != nil {
			err = err0
		}
	}
	return err
}

func markOnLoad(ctx *Context, proc *ProcInfo, used map[int]byte) {
	algo := ctx.Config.Service.PinCoresAlgo

	l0 := len(proc.Threads)
	for i := 0; i < l0; i++ {
		thread := proc.Threads[i]

		// check what cpu from load mask only
		if isMaskInSet(thread.CpuSet, ctx.pool.LoadMask) {
			count := thread.CpuSet.Count()
			// checks cpu count for validity
			if !isLoadCpuCountValid(algo, count) {
				continue
			}

			// mark cpu as used
			MaskIntoMap(thread.CpuSet, used)
		}
	}
}

func pinOnIdle(ctx *Context, proc *ProcInfo) {

}

func pinOnLoad(ctx *Context, proc *ProcInfo) {
	//selection := ctx.Config.Service.Selection
	//patterns := selection.Patterns
	//
	//l0 := len(proc.Threads)
	//for i := 0; i < l0; i++ {
	//	thread := proc.Threads[i]
	//
	//	isSelected := isThreadSelected(thread, patterns)
	//}
}
