package monitoring

type State struct {
	IdleLoad0 float64
	IdleLoad1 float64
	LoadFree0 float64
	LoadFree1 float64
	LoadUsed0 float64
	LoadUsed1 float64
}

func NewState() *State {
	return &State{}
}
