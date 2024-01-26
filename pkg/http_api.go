package pkg

import "pind/pkg/http_api"

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

	state.Error = ""
	return state
}
