package numa

import (
	"github.com/pnsafonov/pind/pkg/utils/core_utils"
	log "github.com/sirupsen/logrus"
	"sort"
)

// PhysCpu - list of physical cores
type PhysCpu struct {
	CoresTopology []*CpuTopologyInfo
	Cores         []*PhysCore
}

// PhysCore - information about physical core topology
type PhysCore struct {
	Id             int              // core id, cpu id
	TopologyInfo   *CpuTopologyInfo // core's topology
	ThreadSiblings []int            // hyper threading cores on same physical core
}

func GetPhysTopology() (*PhysCpu, error) {
	infos, err := GetCpuTopologyInfo()
	if err != nil {
		log.Errorf("GetPhysTopology, GetCpuTopologyInfo err = %v", err)
		return nil, err
	}

	l0 := len(infos)
	cores := make([]*PhysCore, 0, l0)
	for i := 0; i < l0; i++ {
		info := infos[i]

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
		cores = append(cores, core)
	}

	sort.Slice(cores, func(i, j int) bool {
		return cores[i].Id < cores[j].Id
	})

	result := &PhysCpu{
		CoresTopology: infos,
		Cores:         cores,
	}
	return result, nil
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
