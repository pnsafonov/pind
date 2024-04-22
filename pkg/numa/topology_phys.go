package numa

import (
	"github.com/pnsafonov/pind/pkg/utils/core_utils"
	log "github.com/sirupsen/logrus"
	"sort"
)

// NodePhysInfo - information about numa, physical info (topology)
type NodePhysInfo struct {
	Cores []*PhysCore //
	Index int         // numa index
}

// PhysCore - information about physical core topology
type PhysCore struct {
	Id             int              // core id, cpu id
	TopologyInfo   *CpuTopologyInfo // core's topology
	ThreadSiblings []int            // hyper threading cores on same physical core
}

func GetNodesPhysInfo() ([]*NodePhysInfo, error) {
	infos, err := GetCpuTopologyInfo()
	if err != nil {
		log.Errorf("GetPhysTopology, GetCpuTopologyInfo err = %v", err)
		return nil, err
	}

	nodes0 := topologyToPhysInfo(infos)

	l0 := len(infos)
	for i := 0; i < l0; i++ {
		info := infos[i]
		numaId := info.Topology.PhysicalPackageID

		nodeInfo, ok := GetNodePhysInfo0(nodes0, numaId)
		if !ok {
			log.Errorf("GetNodePhysInfo, numa not found, numaId = %d", numaId)
			continue
		}

		cores := nodeInfo.Cores

		sl0 := info.Topology.ThreadSiblingsList
		if IsPhysCoresContains(cores, sl0) {
			// phys core already added
			continue
		}

		threadSiblings := core_utils.CopyIntSlice(sl0)
		core := &PhysCore{
			Id:             info.Number,
			TopologyInfo:   info,
			ThreadSiblings: threadSiblings,
		}
		nodeInfo.Cores = append(cores, core)
	}

	for _, info := range nodes0 {
		cores := info.Cores
		sort.Slice(cores, func(i, j int) bool {
			return cores[i].Id < cores[j].Id
		})
	}

	return nodes0, nil
}

func topologyToPhysInfo(infos []*CpuTopologyInfo) []*NodePhysInfo {
	nodes0 := make([]*NodePhysInfo, 0, 2)

	for _, info := range infos {
		numaId := info.Topology.PhysicalPackageID

		_, ok := GetNodePhysInfo0(nodes0, numaId)
		if !ok {
			node := &NodePhysInfo{
				Index: numaId,
			}
			nodes0 = append(nodes0, node)
		}
	}

	return nodes0
}

func GetNodePhysInfo0(nodes0 []*NodePhysInfo, numaId int) (*NodePhysInfo, bool) {
	for _, node := range nodes0 {
		if node.Index == numaId {
			return node, true
		}
	}
	return nil, false
}

// IsPhysCoresContains - is physical core contains thread siblings cores
func IsPhysCoresContains(cores []*PhysCore, threadSiblings []int) bool {
	for _, core := range cores {
		if core_utils.IsIntSliceEqual(core.ThreadSiblings, threadSiblings) {
			return true
		}
	}
	return false
}
