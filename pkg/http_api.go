package pkg

import (
	"fmt"
	"pind/pkg/http_api"
	"pind/pkg/utils/math_utils"
	"sort"
)

func setHttpApiData(ctx *Context) error {
	state0 := fillHttpApiState(ctx)
	return ctx.HttpApi.SetState(state0)
}

func fillHttpApiState(ctx *Context) *http_api.State {
	cpuInfo := ctx.lastCpuInfo
	state := &http_api.State{}

	for cpu, _ := range ctx.state.Used {
		state.Pool.Load.Used = append(state.Pool.Load.Used, cpu)
	}
	for _, cpu := range ctx.Config.Service.Pool.Load.Values {
		_, ok := ctx.state.Used[cpu]
		if ok {
			continue
		}
		state.Pool.Load.Free = append(state.Pool.Load.Free, cpu)
	}
	state.Pool.Idle = ctx.Config.Service.Pool.Idle.Values
	state.Pool.IdleLoad0 = math_utils.Round2(ctx.pool.IdleLoad0)
	state.Pool.IdleLoad1 = math_utils.Round2(ctx.pool.IdleLoad1)

	state.Time = getProcTimeStr(ctx)

	procs := make(map[string]*http_api.Proc)
	for _, procInfo0 := range ctx.lastAll {
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

		procs[procKey] = proc
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

	state.Error = ""
	return state
}

func getProcTimeStr(ctx *Context) string {
	for _, procInfo0 := range ctx.lastAll {
		return procInfo0.time.String()
	}
	return ""
}
