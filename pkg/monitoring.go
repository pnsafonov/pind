package pkg

import "github.com/pnsafonov/pind/pkg/monitoring"

func setMonitoringState(ctx *Context) error {
	state := fillMonitoringState(ctx)
	ctx.Monitoring.SetState(state)
	return nil
}

func fillMonitoringState(ctx *Context) *monitoring.State {
	state := monitoring.NewState()

	state.Pool.IdleLoad0 = ctx.pool.IdleLoad0
	state.Pool.IdleLoad1 = ctx.pool.IdleLoad1
	state.Pool.LoadFree0 = ctx.pool.LoadFree0
	state.Pool.LoadFree1 = ctx.pool.LoadFree1
	state.Pool.LoadUsed0 = ctx.pool.LoadUsed0
	state.Pool.LoadUsed1 = ctx.pool.LoadUsed1

	l0 := len(ctx.pool.Nodes)
	for i := 0; i < l0; i++ {
		node := ctx.pool.Nodes[i]

		nodeMon := &monitoring.PoolNode{
			Index:     i,
			LoadFree0: node.LoadFree0,
			LoadFree1: node.LoadFree1,
			LoadUsed0: node.LoadUsed0,
			LoadUsed1: node.LoadUsed1,
		}
		state.Pool.Nodes = append(state.Pool.Nodes, nodeMon)
	}

	return state
}
