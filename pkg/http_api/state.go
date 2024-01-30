package http_api

type State struct {
	Error string           `json:"error"`
	Time  string           `json:"time"`
	Procs map[string]*Proc `json:"procs"`
	Pool  Pool             `json:"pool"`
	Numa  []*Numa          `json:"numa"`
}

type Proc struct {
	PID      int      `json:"pid"`
	Comm     string   `json:"comm"`
	VmName   string   `json:"vmname"`
	CPU      float64  `json:"cpu"`
	Load     bool     `json:"load"`
	Filtered bool     `json:"filtered"`
	Threads  []Thread `json:"threads"`
}

type Thread struct {
	PID      int    `json:"pid"`
	Comm     string `json:"comm"`
	Ignored  bool   `json:"ignored"`
	Selected bool   `json:"selected"`
	Cores    Cores  `json:"cores"`
}

type Cores struct {
	Assigned []int `json:"assigned"`
	Actual   []int `json:"actual"`
}

type Pool struct {
	IdleLoad0 float64 `json:"idle_load0"`
	IdleLoad1 float64 `json:"idle_load1"`
	Idle      []int   `json:"idle"`
	Load      Load    `json:"load"`
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
