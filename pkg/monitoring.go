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

	return state
}
