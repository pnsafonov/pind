package pkg

import (
	"fmt"
	"github.com/pnsafonov/pind/pkg/http_api"
	"github.com/pnsafonov/pind/pkg/utils/core_utils"
	"github.com/pnsafonov/pind/pkg/utils/math_utils"
	"sort"
	"time"
)

func setHttpApiData(ctx *Context) error {
	state0 := fillHttpApiState(ctx)
	return ctx.HttpApi.SetState(state0)
}

func fillHttpApiState(ctx *Context) *http_api.State {
	cpuInfo := ctx.lastCpuInfo
	pool := ctx.pool
	state := &http_api.State{}

	state.Version = ctx.Version
	state.GitHash = ctx.GitHash

	l1 := len(pool.Nodes)
	nodes := make([]*http_api.PoolNode, 0, l1)
	for i := 0; i < l1; i++ {
		node0 := pool.Nodes[i]
		free := node0.getLoadFreeSlice()
		used := node0.getLoadUsedSlice()
		node := &http_api.PoolNode{
			Index:    node0.Index,
			LoadFree: free,
			LoadUsed: used,
		}
		nodes = append(nodes, node)
	}
	pool0 := &http_api.Pool{
		Idle:      ctx.Config.Service.Pool.Idle.Values,
		IdleLoad0: math_utils.Round2(ctx.pool.IdleLoad0),
		IdleLoad1: math_utils.Round2(ctx.pool.IdleLoad1),
		Nodes:     nodes,
		LoadType:  ctx.Config.Service.Pool.LoadType,
	}
	state.Pool = pool0

	state.Time = getProcTimeStr(ctx)

	notInFilter := getProcsNotInFilter(ctx)
	inFilter := getProcsInFilter(ctx)
	procs := &http_api.Procs{
		NotInFilter: notInFilter,
		InFilter:    inFilter,
	}
	state.Procs = procs

	l0 := len(cpuInfo.Infos)
	for i := 0; i < l0; i++ {
		node := cpuInfo.Infos[i]
		numa0 := &http_api.Numa{
			Index: i,
		}

		l1 := len(node.Cores)
		numa0.Cpus = make([]*http_api.CPU, 0, l1)
		for cpu, cpuInfo0 := range node.Cores {
			cpuLoad0 := math_utils.Round2(cpuInfo0.CpuLoad)
			cpu0 := &http_api.CPU{
				Index: cpu,
				Load:  cpuLoad0,
			}
			numa0.Cpus = append(numa0.Cpus, cpu0)
		}
		state.Numa = append(state.Numa, numa0)
		sort.Slice(numa0.Cpus, func(i, j int) bool {
			return numa0.Cpus[i].Index < numa0.Cpus[j].Index
		})
	}
	sort.Slice(state.Numa, func(i, j int) bool {
		return state.Numa[i].Index < state.Numa[j].Index
	})

	errs := getErrors(ctx)
	state.Errors = errs

	config0 := &http_api.Config{
		Filters0:  ctx.Config.Service.Filters0,
		Filters1:  ctx.Config.Service.Filters1,
		Selection: ctx.Config.Service.Selection,
		Ignore:    ctx.Config.Service.Ignore,
	}
	state.Config = config0

	return state
}

func mapToPoolCoresList(map0 map[int]*PoolCore) []*http_api.PoolCore {
	l0 := len(map0)
	result := make([]*http_api.PoolCore, 0, l0)
	for _, poolCore := range map0 {
		poolCore0 := &http_api.PoolCore{
			Id:        poolCore.Id,
			Available: core_utils.CopyIntSlice(poolCore.Available),
			Used:      core_utils.CopyIntSlice(poolCore.Used),
		}
		result = append(result, poolCore0)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Id < result[j].Id
	})
	return result
}

func getProcTimeStr(ctx *Context) string {
	for _, procInfo0 := range ctx.lastAll {
		return procInfo0.time.Format(time.RFC3339)
	}
	return ""
}

func getProcsNotInFilter(ctx *Context) map[string]*http_api.Proc {
	procs := make(map[string]*http_api.Proc)
	for _, procInfo0 := range ctx.lastNotInFilter {
		cpu0 := math_utils.Round2(procInfo0.cpu0)
		proc := &http_api.Proc{
			PID:  procInfo0.Stat.PID,
			Comm: procInfo0.Stat.Comm,
			CPU:  cpu0,
			Load: procInfo0.load,
		}

		procKey := ""
		vmName, ok := parseVmName(procInfo0.Cmd)
		if ok {
			proc.VmName = vmName
			procKey = fmt.Sprintf("%s %d", vmName, procInfo0.Stat.PID)
		} else {
			procKey = fmt.Sprintf("%d", procInfo0.Stat.PID)
		}

		l2 := len(procInfo0.Threads)
		proc.Threads = make([]*http_api.Thread, 0, l2)
		for i := 0; i < l2; i++ {
			thread := procInfo0.Threads[i]

			coresActual := MaskToArray(&thread.CpuSet)
			cores := http_api.Cores{
				Assigned: nil,
				Actual:   coresActual,
			}

			thread0 := &http_api.Thread{
				PID:     thread.Stat.PID,
				Comm:    thread.Stat.Comm,
				Ignored: thread.Ignored,
				Cores:   cores,
			}

			proc.Threads = append(proc.Threads, thread0)
		}

		procs[procKey] = proc
	}
	return procs
}

func getProcsInFilter(ctx *Context) map[string]*http_api.Proc {
	procs := make(map[string]*http_api.Proc)
	for _, procInfo1 := range ctx.state.Procs {
		procInfo0 := procInfo1.ProcInfo

		notSelectedCores := procInfo1.NotSelected.getCpusCopy0()
		cpu0 := math_utils.Round2(procInfo0.cpu0)
		proc := &http_api.Proc{
			PID:              procInfo0.Stat.PID,
			Comm:             procInfo0.Stat.Comm,
			CPU:              cpu0,
			Load:             procInfo0.load,
			NotSelectedCores: notSelectedCores,
		}

		procKey := ""
		vmName, ok := parseVmName(procInfo0.Cmd)
		if ok {
			proc.VmName = vmName
			procKey = fmt.Sprintf("%s %d", vmName, procInfo0.Stat.PID)
		} else {
			procKey = fmt.Sprintf("%d", procInfo0.Stat.PID)
		}

		l2 := len(procInfo1.Threads)
		proc.Threads = make([]*http_api.Thread, 0, l2)
		for _, thread1 := range procInfo1.Threads {
			thread := thread1.ThreadInfo
			selected0 := ThreadSelectionToBool0(thread1.Selected)

			coresAssigned := thread1.Cpus.getCpusCopy0()
			coresActual := MaskToArray(&thread.CpuSet)
			cores := http_api.Cores{
				Assigned: coresAssigned,
				Actual:   coresActual,
			}

			thread0 := &http_api.Thread{
				PID:      thread.Stat.PID,
				Comm:     thread.Stat.Comm,
				Ignored:  thread.Ignored,
				Selected: selected0,
				Cores:    cores,
			}

			proc.Threads = append(proc.Threads, thread0)
		}

		procs[procKey] = proc
	}
	return procs
}

func ThreadSelectionToBool0(val ThreadSelection) (result *bool) {
	if val == ThreadSelectionYes {
		result = new(bool)
		*result = true
		return
	}
	if val == ThreadSelectionNo {
		result = new(bool)
		*result = false
		return
	}
	return
}

func getErrors(ctx *Context) *http_api.Errors {
	errs0 := ctx.state.Errors

	errs := &http_api.Errors{
		CalcProcsCPU:         getErrorStr(errs0.CalcCoresCPU),
		CalcCoresCPU:         getErrorStr(errs0.CalcCoresCPU),
		PinNotInFilterToIdle: getErrorStr(errs0.PinNotInFilterToIdle),
		StatePinIdle:         getErrorStr(errs0.StatePinIdle),
		StatePinLoad:         getErrorStr(errs0.StatePinLoad),
		IdleOverwork:         getErrorStr(errs0.IdleOverwork),
	}
	errs.RequiredCPU.Total = errs0.RequiredCPU.Total
	errs.RequiredCPU.PerProcess = core_utils.CopyIntSlice(errs0.RequiredCPU.PerProcess)

	return errs
}

func getErrorStr(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
