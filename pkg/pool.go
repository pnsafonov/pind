package pkg

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"pind/pkg/config"
	"pind/pkg/numa"
)

type Pool struct {
	Config config.Pool
	Nodes  []*PoolNodeInfo

	//FullMask unix.CPUSet
	//IdleMask unix.CPUSet
	//LoadMask unix.CPUSet

	IdleLoadFull0 float64 // 400, 600, 800 %
	IdleLoad0     float64 // 400, 600, 800 %
	IdleLoad1     float64 // 0-100 %
}

type PoolNodeInfo struct {
	Index    int
	Node     *numa.NodeInfo
	LoadFree map[int]byte
	LoadUsed map[int]byte
}

func NewPoolNodes(nodes []*numa.NodeInfo, config0 config.Pool) []*PoolNodeInfo {
	l0 := len(nodes)
	poolNodes := make([]*PoolNodeInfo, 0, l0)
	for i := 0; i < l0; i++ {
		node := nodes[i]
		free := make(map[int]byte)
		poolNode := &PoolNodeInfo{
			Index:    node.Index,
			Node:     node,
			LoadFree: free,
			LoadUsed: make(map[int]byte),
		}
		poolNodes = append(poolNodes, poolNode)

		for _, cpu := range node.Cpus {
			if config.IsCpuInInterval(cpu, config0.Load) {
				free[cpu] = 1
			}
		}
	}

	return poolNodes
}

func NewPool(config0 config.Pool) (*Pool, error) {
	nodes, err := numa.GetNodes()
	if err != nil {
		log.Errorf("NewPool, numa.GetNodes err = %v", err)
		return nil, err
	}

	poolNodes := NewPoolNodes(nodes, config0)

	//fullMask := numa.NodesToFullMask(nodes)
	//idleMask := numa.CpusToMask(config0.Idle.Values)
	//loadMask := numa.CpusToMask(config0.Load.Values)

	l0 := len(config0.Idle.Values)
	idleLoadFull := float64(l0) * 100

	pool := &Pool{
		Config: config0,
		Nodes:  poolNodes,
		//FullMask:      fullMask,
		//IdleMask:      idleMask,
		//LoadMask:      loadMask,
		IdleLoadFull0: idleLoadFull,
	}

	return pool, nil
}

// isMasksEqual - if masks are equal
func isMasksEqual(mask0 unix.CPUSet, mask1 unix.CPUSet) bool {
	l0 := len(mask0)
	for i := 0; i < l0; i++ {
		if mask0[i] != mask1[i] {
			return false
		}
	}
	return true
}

// isMaskInSet - if all mask's bits in the set
func isMaskInSet(mask unix.CPUSet, set unix.CPUSet) bool {
	l0 := len(mask)
	for i := 0; i < l0; i++ {
		if mask[i]|set[i] != set[i] {
			return false
		}
	}
	return true
}

// MaskIntoMap - fill map with mask cpu values
func MaskIntoMap(mask unix.CPUSet, map0 map[int]byte) {
	l0 := len(mask)
	for i := 0; i < l0; i++ {
		m0 := uint64(mask[i])
		if m0 == 0 {
			// most of 16 mask values are 0
			continue
		}

		num0 := i * 64
		for j := 0; j < 64; j++ {
			v0 := uint64(1) << j
			if m0&v0 != 0 {
				cpu := num0 + j
				map0[cpu] = 1
			}
		}
	}
}

func IsMaskNotZero(mask unix.CPUSet) bool {
	l0 := len(mask)
	for i := 0; i < l0; i++ {
		m0 := uint64(mask[i])
		if m0 != 0 {
			return true
		}
	}
	return false
}

func ZeroMask(mask unix.CPUSet) {
	isNotZero := IsMaskNotZero(mask)
	if isNotZero {
		ZeroMask0(mask)
	}
}

func ZeroMask0(mask unix.CPUSet) {
	l0 := len(mask)
	for i := 0; i < l0; i++ {
		m0 := uint64(mask[i])
		if m0 != 0 {
			m0 = 0
		}
	}
}

func MaskToArray(mask *unix.CPUSet) []int {
	cpus := make([]int, 0, 4)

	l0 := len(mask)
	for i := 0; i < l0; i++ {
		m0 := uint64(mask[i])
		if m0 == 0 {
			// most of 16 mask values are 0
			continue
		}

		num0 := i * 64
		for j := 0; j < 64; j++ {
			v0 := uint64(1) << j
			if m0&v0 != 0 {
				cpu := num0 + j
				cpus = append(cpus, cpu)
			}
		}
	}

	return cpus
}

// isLoadCpuCountValid - checks is cpu count valid in mask
// of thread on load
func isLoadCpuCountValid(algo *config.PinCoresAlgo, count int) bool {
	return count == algo.Selected || count == algo.NotSelected
}
