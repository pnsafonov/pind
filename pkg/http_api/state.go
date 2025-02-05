package http_api

import "github.com/pnsafonov/pind/pkg/config"

type State struct {
	Version string  `json:"version"`
	GitHash string  `json:"git_hash"`
	Errors  *Errors `json:"errors"`
	Time    string  `json:"time"`
	Procs   *Procs  `json:"procs"`

	Pool *Pool   `json:"pool"`
	Numa []*Numa `json:"numa"`

	Config *Config `json:"config"`
}

type Procs struct {
	NotInFilter map[string]*Proc `json:"not_in_filter"`
	InFilter    map[string]*Proc `json:"in_filter"`
}

type Config struct {
	Filters0  []*config.ProcFilter `json:"filters0"`
	Filters1  []*config.ProcFilter `json:"filters1"`
	Selection config.Selection     `json:"selection"`
	Ignore    *config.Ignore       `json:"ignore"`
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
	IdleLoad0    float64     `json:"idle_load0"`
	IdleLoad1    float64     `json:"idle_load1"`
	IdleLoadFull float64     `json:"idle_load_full"`
	Idle         []int       `json:"idle"`
	Nodes        []*PoolNode `json:"numa_nodes"`
	LoadType     string      `json:"load_type"`
	LoadMode     string      `json:"load_mode"`
	LoadFree0    float64     `json:"load_free0"`
	LoadFree1    float64     `json:"load_free1"`
	LoadFreeFull float64     `json:"load_free_full"`
	LoadUsed0    float64     `json:"load_used0"`
	LoadUsed1    float64     `json:"load_used1"`
	LoadUsedFull float64     `json:"load_used_full"`
}

type PoolNode struct {
	Index        int         `json:"index"`
	LoadFree     []*PoolCore `json:"load_free"`
	LoadUsed     []*PoolCore `json:"load_used"`
	LoadFree0    float64     `json:"load_free0"`
	LoadFree1    float64     `json:"load_free1"`
	LoadFreeFull float64     `json:"load_free_full"`
	LoadUsed0    float64     `json:"load_used0"`
	LoadUsed1    float64     `json:"load_used1"`
	LoadUsedFull float64     `json:"load_used_full"`
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
