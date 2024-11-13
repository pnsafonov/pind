package http_api

import "strings"

func (x *State) CloneWithFilter0(inFilter map[string]*Proc) *State {
	procs0 := &Procs{
		NotInFilter: x.Procs.NotInFilter,
		InFilter:    inFilter,
	}

	state0 := &State{
		Version: x.Version,
		GitHash: x.GitHash,
		Errors:  x.Errors,
		Time:    x.Time,
		Procs:   procs0,

		Pool: x.Pool,
		Numa: x.Numa,

		Config: x.Config,
	}

	return state0
}

func (x *State) CloneWithFilter1(vmName string) *State {
	inFilter := make(map[string]*Proc)
	for k, v := range x.Procs.InFilter {
		if v.VmName == vmName {
			inFilter[k] = v
		}
	}

	return x.CloneWithFilter0(inFilter)
}

func (x *State) CloneWithFilter2(vmPrefix string) *State {
	inFilter := make(map[string]*Proc)
	for k, v := range x.Procs.InFilter {
		if strings.HasPrefix(v.VmName, vmPrefix) {
			inFilter[k] = v
		}
	}

	return x.CloneWithFilter0(inFilter)
}
