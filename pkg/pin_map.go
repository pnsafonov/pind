package pkg

import "sync"

// PinMap - direct order for process mapping
type PinMap struct {
	mu      sync.Mutex
	vmNames map[string]int
}

func NewPinMap() *PinMap {
	map0 := &PinMap{
		vmNames: make(map[string]int),
	}
	return map0
}

func (x *PinMap) GetNumaNode(vmName string) (int, bool) {
	x.mu.Lock()
	defer x.mu.Unlock()

	if vmName == "" {
		return -1, false
	}

	nodeId, ok := x.vmNames[vmName]
	if !ok {
		return -1, false
	}

	return nodeId, ok
}

func (x *PinMap) AddPinMapping(vmName string, nodeId int) {
	x.mu.Lock()
	defer x.mu.Unlock()

	x.vmNames[vmName] = nodeId
}

func (x *PinMap) Clone0() map[string]int {
	m0 := make(map[string]int)

	x.mu.Lock()
	defer x.mu.Unlock()

	for k, v := range x.vmNames {
		m0[k] = v
	}

	return m0
}

func (x *PinMap) Remove(vmName string) {
	x.mu.Lock()
	defer x.mu.Unlock()

	delete(x.vmNames, vmName)
}

func (x *PinMap) Clean() {
	x.mu.Lock()
	defer x.mu.Unlock()

	x.vmNames = make(map[string]int)
}
