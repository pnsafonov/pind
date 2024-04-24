package numa

import (
	"fmt"
	"github.com/lrita/numa"
	"github.com/prometheus/procfs"
	"github.com/prometheus/procfs/sysfs"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

var (
	initError           error
	nodeMaxId           = -1
	nodesCount          = 0
	nodes               []*NodeInfo
	nodesPhys           []*NodePhysInfo
	procfs0             procfs.FS
	ErrNumaNodeNotFound = fmt.Errorf("numa_node_not_found")
	ErrNotSameNumaNodes = fmt.Errorf("not_same_numa_nodes")

	sysfs0 sysfs.FS
)

func init() {
	available := numa.Available()
	if !available {
		initError = fmt.Errorf("numa_not_availble")
		return
	}

	nodeMaxId = numa.MaxNodeID()
	nodesCount = nodeMaxId + 1

	nodes, initError = initNodes(nodesCount)
	if initError != nil {
		return
	}

	procfs0, initError = procfs.NewDefaultFS()
	if initError != nil {
		return
	}

	sysfs0, initError = sysfs.NewDefaultFS()
	if initError != nil {
		return
	}

	nodesPhys, initError = GetNodesPhysInfo()
	if initError != nil {
		return
	}

	l0 := len(nodes)
	l1 := len(nodesPhys)
	if l0 != l1 {
		initError = fmt.Errorf("count of nodes and nodesPhys is not same, %d != %d", l0, l1)
	}
}

func initNodes(nodesCount int) ([]*NodeInfo, error) {
	result := make([]*NodeInfo, 0, nodesCount)
	for i := 0; i < nodesCount; i++ {
		mask0, err := numa.NodeToCPUMask(i)
		if err != nil {
			return nil, err
		}
		ni, err := maskToNodeInfo(mask0)
		if err != nil {
			return nil, err
		}
		ni.Index = i
		result = append(result, ni)
	}
	return result, nil
}

type NodeInfo struct {
	Mask  []uint64 // Bitmask, CPUSet, unix.CPUSet
	Cpus  []int    //
	Index int      // numa index
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
	nodes0, err := GetNodes()
	if err != nil {
		return err
	}
	l0 := len(nodes0)
	for i := 0; i < l0; i++ {
		ni := nodes0[i]
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

func PrintNumaPhys0() error {
	nodes0, err := GetNodesPhys()
	if err != nil {
		return err
	}
	l0 := len(nodes0)
	for i := 0; i < l0; i++ {
		ni := nodes0[i]
		_, _ = fmt.Printf("numa %d\n", ni.Index)

		l1 := len(ni.Cores)
		for j := 0; j < l1; j++ {
			core := ni.Cores[j]
			_, _ = fmt.Printf("phys core %d, siblings = %v\n", core.Id, core.ThreadSiblings)
		}
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

func GetNodesPhys() ([]*NodePhysInfo, error) {
	if initError != nil {
		return nil, initError
	}
	return nodesPhys, nil
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

func getNumaNode(cpu int, nodes0 []*NodeInfo) (*NodeInfo, error) {
	for _, node := range nodes0 {
		for _, cpuNode := range node.Cpus {
			if cpu == cpuNode {
				return node, nil
			}
		}
	}
	return nil, ErrNumaNodeNotFound
}

func IsCpusOnSameNumaNode(cpus []int) error {
	nodes0, err := GetNodes()
	if err != nil {
		log.Errorf("IsCpuOnOneNumaNode, GetNodes err = %v", err)
		return err
	}

	var sameNode *NodeInfo
	for _, cpu := range cpus {
		node, err0 := getNumaNode(cpu, nodes0)
		if err0 != nil {
			return err0
		}
		if sameNode == nil {
			sameNode = node
			continue
		}
		if sameNode != node {
			return ErrNotSameNumaNodes
		}
	}

	return nil
}
