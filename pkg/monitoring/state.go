package monitoring

type State struct {
	IdleLoad0 float64
	IdleLoad1 float64
}

func NewState() *State {
	return &State{}
}
