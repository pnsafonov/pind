package pkg

import (
	"fmt"
	"pind/pkg/http_api"
)

func setHttpApiData(ctx *Context) error {
	state0 := fillHttpApiState(ctx)
	return ctx.HttpApi.SetState(state0)
}

func fillHttpApiState(ctx *Context) *http_api.State {
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

	state.Time = getProcTimeStr(ctx)

	procs := make(map[string]*http_api.Proc)
	for _, procInfo0 := range ctx.lastAll {
		proc := &http_api.Proc{
			PID:  procInfo0.Stat.PID,
			Comm: procInfo0.Stat.Comm,
			CPU:  procInfo0.cpu0,
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

	state.Error = ""
	return state
}

func getProcTimeStr(ctx *Context) string {
	for _, procInfo0 := range ctx.lastAll {
		return procInfo0.time.String()
	}
	return ""
}
