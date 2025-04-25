package mon_state

import "time"

type State struct {
	Time time.Time
	Pool *Pool
	//Procs []*Proc
	Procs map[string]*Proc
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
	Index         int
	LoadFree0     float64
	LoadFree1     float64
	LoadUsed0     float64
	LoadUsed1     float64
	LoadFreeCount float64
	LoadUsedCount float64
}

type Proc struct {
	VmName            string
	Time              time.Time
	CPU               float64
	Load              bool
	Numa0             int
	RequiredCoresPhys int
	RequiredCores     int
	AssignedCores     int
}

func NewState() *State {
	now := time.Now()
	pool := &Pool{}

	return &State{
		Time:  now,
		Pool:  pool,
		Procs: make(map[string]*Proc),
	}
}

func (x *Proc) GetLoad0() float64 {
	if x.Load {
		return 1
	}
	return 0
}
