package http_api

type State struct {
	Errors *Errors `json:"errors"`
	Time   string  `json:"time"`
	Procs  *Procs  `json:"procs"`

	Pool *Pool   `json:"pool"`
	Numa []*Numa `json:"numa"`
}

type Procs struct {
	NotInFilter map[string]*Proc `json:"not_in_filter"`
	InFilter    map[string]*Proc `json:"in_filter"`
}

type Proc struct {
	PID              int       `json:"pid"`
	Comm             string    `json:"comm"`
	VmName           string    `json:"vmname"`
	CPU              float64   `json:"cpu"`
	Load             bool      `json:"load"`
	NotSelectedCores *[]int    `json:"not_selected_cores,omitempty"`
	Threads          []*Thread `json:"threads"`
}

type Thread struct {
	PID      int    `json:"pid"`
	Comm     string `json:"comm"`
	Ignored  bool   `json:"ignored"`
	Cores    Cores  `json:"cores"`
	Selected *bool  `json:"selected,omitempty"`
}

type Cores struct {
	Assigned *[]int `json:"assigned,omitempty"`
	Actual   []int  `json:"actual"`
}

type Pool struct {
	IdleLoad0 float64     `json:"idle_load0"`
	IdleLoad1 float64     `json:"idle_load1"`
	Idle      []int       `json:"idle"`
	Nodes     []*PoolNode `json:"numa_nodes"`
	LoadType  string      `json:"load_type"`
}

type PoolNode struct {
	Index    int         `json:"index"`
	LoadFree []*PoolCore `json:"load_free"`
	LoadUsed []*PoolCore `json:"load_used"`
}

type PoolCore struct {
	Id        int   `json:"id"`
	Available []int `json:"available,omitempty"`
	Used      []int `json:"used,omitempty"`
}

type Load struct {
	Used []int `json:"used"`
	Free []int `json:"free"`
}

type Numa struct {
	Index int    `json:"index"`
	Cpus  []*CPU `json:"cpus"`
}

type CPU struct {
	Index int     `json:"index"`
	Load  float64 `json:"load"`
}

type Errors struct {
	CalcProcsCPU         string      `json:"calc_procs_cpu"`
	CalcCoresCPU         string      `json:"calc_cores_cpu"`
	PinNotInFilterToIdle string      `json:"pin_not_in_filter_to_idle"`
	StatePinIdle         string      `json:"state_pin_idle"`
	StatePinLoad         string      `json:"state_pin_load"`
	IdleOverwork         string      `json:"idle_overwork"`
	RequiredCPU          RequiredCPU `json:"required_cpu"`
}

type RequiredCPU struct {
	Total      int   `json:"total"`
	PerProcess []int `json:"per_process"`
}
