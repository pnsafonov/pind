package monitoring

type State struct {
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
	pool := &Pool{}

	return &State{
		pool,
	}
}
