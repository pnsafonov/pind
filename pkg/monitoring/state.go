package monitoring

import "time"

type State struct {
	Time time.Time
	Pool *Pool
}

type Pool struct {
	IdleLoad0 float64
	IdleLoad1 float64
	LoadFree0 float64
	LoadFree1 float64
	LoadUsed0 float64
	LoadUsed1 float64

	Nodes []*PoolNode
}

type PoolNode struct {
	Index     int
	LoadFree0 float64
	LoadFree1 float64
	LoadUsed0 float64
	LoadUsed1 float64
}

func NewState() *State {
	now := time.Now()
	pool := &Pool{}

	return &State{
		Time: now,
		Pool: pool,
	}
}
