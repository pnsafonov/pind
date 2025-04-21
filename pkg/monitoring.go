package pkg

import "github.com/pnsafonov/pind/pkg/monitoring"

func setMonitoringState(ctx *Context) error {
	state := fillMonitoringState(ctx)
	ctx.Monitoring.SetState(state)
	return nil
}

func fillMonitoringState(ctx *Context) *monitoring.State {
	state := monitoring.NewState()

	state.IdleLoad0 = ctx.pool.IdleLoad0
	state.IdleLoad1 = ctx.pool.IdleLoad1
	state.LoadFree0 = ctx.pool.LoadFree0
	state.LoadFree1 = ctx.pool.LoadFree1
	state.LoadUsed0 = ctx.pool.LoadUsed0
	state.LoadUsed1 = ctx.pool.LoadUsed1

	return state
}
