package pkg

import (
	"github.com/pnsafonov/pind/pkg/monitoring/mon_state"
)

func setMonitoringState(ctx *Context) error {
	state := fillMonitoringState(ctx)
	ctx.Monitoring.SetState(state)
	return nil
}

func fillMonitoringState(ctx *Context) *mon_state.State {
	state := mon_state.NewState()

	state.Pool.IdleLoad0 = ctx.pool.IdleLoad0
	state.Pool.IdleLoad1 = ctx.pool.IdleLoad1
	state.Pool.LoadFree0 = ctx.pool.LoadFree0
	state.Pool.LoadFree1 = ctx.pool.LoadFree1
	state.Pool.LoadUsed0 = ctx.pool.LoadUsed0
	state.Pool.LoadUsed1 = ctx.pool.LoadUsed1

	l0 := len(ctx.pool.Nodes)
	for i := 0; i < l0; i++ {
		node := ctx.pool.Nodes[i]

		l1 := len(node.LoadFree)
		l2 := len(node.LoadUsed)
		nodeMon := &mon_state.PoolNode{
			Index:         i,
			LoadFree0:     node.LoadFree0,
			LoadFree1:     node.LoadFree1,
			LoadUsed0:     node.LoadUsed0,
			LoadUsed1:     node.LoadUsed1,
			LoadFreeCount: float64(l1),
			LoadUsedCount: float64(l2),
		}
		state.Pool.Nodes = append(state.Pool.Nodes, nodeMon)
	}

	for _, pinProc := range ctx.state.Procs {
		if pinProc.ProcInfo.VmName == "" {
			continue
		}

		numa0 := pinProc.GetNuma0()
		proc := &mon_state.Proc{
			VmName:            pinProc.ProcInfo.VmName,
			Time:              state.Time,
			CPU:               pinProc.ProcInfo.cpu0,
			Load:              pinProc.ProcInfo.load,
			Numa0:             numa0,
			RequiredCoresPhys: pinProc.RequiredCoresPhys,
			RequiredCores:     pinProc.RequiredCores,
			AssignedCores:     pinProc.AssignedCores,
		}

		state.Procs[pinProc.ProcInfo.VmName] = proc
	}

	return state
}
