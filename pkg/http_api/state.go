package http_api

type State struct {
	Error string           `json:"error"`
	Time  string           `json:"time"`
	Procs map[string]*Proc `json:"procs"`
	Pool  Pool             `json:"pool"`
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
	Idle []int `json:"idle"`
	Load Load  `json:"load"`
}

type Load struct {
	Used []int `json:"used"`
	Free []int `json:"free"`
}
