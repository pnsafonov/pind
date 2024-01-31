package http_api

type State struct {
	Error string `json:"error"`
	Time  string `json:"time"`
	Procs *Procs `json:"procs"`

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
	Filtered         bool      `json:"filtered"`
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
}

type PoolNode struct {
	Index    int   `json:"index"`
	LoadFree []int `json:"load_free"`
	LoadUsed []int `json:"load_used"`
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
