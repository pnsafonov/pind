package numa

import (
	"fmt"
	"github.com/lrita/numa"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

var (
	initError  error
	nodeMaxId  = -1
	nodesCount = 0
	nodes      []*NodeInfo
)

func init() {
	available := numa.Available()
	if !available {
		initError = fmt.Errorf("numa_not_availble")
		return
	}

	nodeMaxId = numa.MaxNodeID()
	nodesCount = nodeMaxId + 1

	nodes = make([]*NodeInfo, 0, nodesCount)
	for i := 0; i < nodesCount; i++ {
		mask0, err0 := numa.NodeToCPUMask(i)
		if err0 != nil {
			initError = err0
			return
		}
		ni, err0 := maskToNodeInfo(mask0)
		if err0 != nil {
			initError = err0
			return
		}
		ni.Index = i
		nodes = append(nodes, ni)
	}
}

type NodeInfo struct {
	Mask  []uint64 // Bitmask, CPUSet, unix.CPUSet
	Cpus  []int
	Index int
}

func maskToNodeInfo(mask []uint64) (*NodeInfo, error) {
	l0 := len(mask)
	if l0 > 16 {
		return nil, fmt.Errorf("rocessor_has_too_many_cores")
	}

	// bit mask to cpu index
	cpus := make([]int, 0, 64)
	for i := 0; i < l0; i++ {
		m0 := mask[i]
		num0 := i * 64

		for j := 0; j < 64; j++ {
			v0 := uint64(1) << j
			if m0&v0 != 0 {
				cpu := num0 + j
				cpus = append(cpus, cpu)
			}
		}
	}

	ni := &NodeInfo{
		Mask: mask,
		Cpus: cpus,
	}
	return ni, nil
}

func checkInitError() error {
	if initError != nil {
		log.Printf("Numa is not inited, cause err = %v", initError)
		return initError
	}
	return nil
}

func PrintNuma0() error {
	err := checkInitError()
	if err != nil {
		return err
	}

	l0 := len(nodes)
	for i := 0; i < l0; i++ {
		ni := nodes[i]
		_, _ = fmt.Printf("numa %d\n", ni.Index)

		l1 := len(ni.Mask)
		if l1 > 0 {
			_, _ = fmt.Printf("mask %064b\n", ni.Mask[0])

			for j := 1; j < l1; j++ {
				_, _ = fmt.Printf("     %064b\n", ni.Mask[j])
			}
		}

		_, _ = fmt.Printf("cpus %v\n", ni.Cpus)
		_, _ = fmt.Printf("\n")
	}
	return nil
}

// GetNodes - get information about NUMA
func GetNodes() ([]*NodeInfo, error) {
	if initError != nil {
		return nil, initError
	}
	return nodes, nil
}

func NodesToFullMask(nodes []*NodeInfo) unix.CPUSet {
	mask := unix.CPUSet{}

	l0 := len(nodes)
	for i := 0; i < l0; i++ {
		node := nodes[i]

		l1 := len(node.Cpus)
		for j := 0; j < l1; j++ {
			cpu := node.Cpus[j]
			mask.Set(cpu)
		}
	}

	return mask
}

func CpusToMask(cpus []int) unix.CPUSet {
	mask := unix.CPUSet{}

	l0 := len(cpus)
	for i := 0; i < l0; i++ {
		cpu := cpus[i]
		mask.Set(cpu)
	}

	return mask
}
