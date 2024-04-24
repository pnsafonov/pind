package pkg

import (
	"github.com/pnsafonov/pind/pkg/config"
	"github.com/pnsafonov/pind/pkg/http_api"
	"github.com/pnsafonov/pind/pkg/numa"
	"github.com/pnsafonov/pind/pkg/utils/core_utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"sort"
)

type Pool struct {
	Config    config.Pool
	Nodes     []*PoolNodeInfo
	NodeIndex int

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
	NodePhys *numa.NodePhysInfo
	LoadFree map[int]*PoolCore
	LoadUsed map[int]*PoolCore
	Config   config.Pool
}

// PoolCore - stores information about physical core.
// If core is not physical, PoolCore is not used.
type PoolCore struct {
	Id        int   // core id
	Available []int // Cores that can be used. Like free, but used. Once filled.
	Used      []int // Actually used cores.
}

func NewPoolNodes(nodes []*numa.NodeInfo, nosesPhys []*numa.NodePhysInfo, config0 config.Pool) []*PoolNodeInfo {
	l0 := len(nodes)
	poolNodes := make([]*PoolNodeInfo, 0, l0)
	for i := 0; i < l0; i++ {
		node := nodes[i]
		nodePhys := nosesPhys[i]
		free := make(map[int]*PoolCore)
		poolNode := &PoolNodeInfo{
			Index:    node.Index,
			Node:     node,
			NodePhys: nodePhys,
			LoadFree: free,
			LoadUsed: make(map[int]*PoolCore),
			Config:   config0,
		}
		poolNodes = append(poolNodes, poolNode)

		if config0.LoadType == config.Phys {
			for _, core := range nodePhys.Cores {
				cpu := core.Id
				if config.IsCpuInInterval(cpu, config0.Load) {
					cpus := core_utils.CopyIntSlice(core.ThreadSiblings)
					poolCore := &PoolCore{
						Id:        cpu,
						Available: cpus,
					}
					free[cpu] = poolCore
				}
			}

		} else {
			for _, cpu := range node.Cpus {
				if config.IsCpuInInterval(cpu, config0.Load) {
					poolCore := &PoolCore{
						Id: cpu,
					}
					free[cpu] = poolCore
				}
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

	nodesPhys, err := numa.GetNodesPhys()
	if err != nil {
		log.Errorf("NewPool, numa.GetNodesPhys err = %v", err)
		return nil, err
	}

	poolNodes := NewPoolNodes(nodes, nodesPhys, config0)

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

// getNumaNodeForLoadAssign - search for numa node changing starting node
func (x *Pool) getNumaNodeForLoadAssign(requiredCount int) (*PoolNodeInfo, bool) {
	l0 := len(x.Nodes)
	var freeNode *PoolNodeInfo
	counter := 0
	i := x.NodeIndex
	for {
		if i >= l0 {
			i = 0
		}
		if counter >= l0 {
			break
		}

		node := x.Nodes[i]
		freeCount := len(node.LoadFree)
		if freeCount >= requiredCount {
			freeNode = node
			i++
			break
		}

		i++
		counter++
	}
	x.NodeIndex = i
	return freeNode, freeNode != nil
}

func (x *PoolNodeInfo) assignCores(ctx *Context, proc *PinProc) int {
	node := x
	noSelected := ctx.Config.Service.PinCoresAlgo.NotSelected
	selected := ctx.Config.Service.PinCoresAlgo.Selected

	count := 0
	for _, thread := range proc.Threads {
		if thread.ThreadInfo.Ignored {
			continue
		}
		if thread.Selected == ThreadSelectionNo {
			count += thread.Cpus.AssignRequiredCores1(node, noSelected, &proc.NotSelected)
			continue
		}
		if thread.Selected == ThreadSelectionYes {
			count += thread.Cpus.AssignRequiredCores0(node, selected)
			continue
		}
		log.Warningf("PoolNodeInfo assignCores, execution must not be here!")
	}
	return count
}

func (x *PoolNodeInfo) getFreeCore(physCore int) (int, int, bool) {
	if x.Config.LoadType == config.Phys {
		return x.getFreeCorePhys(physCore)
	}
	core, ok := x.getFreeCoreLogical()
	return -1, core, ok
}

// getFreeCore - returns free core
func (x *PoolNodeInfo) getFreeCoreLogical() (int, bool) {
	for cpu, poolCore := range x.LoadFree {
		delete(x.LoadFree, cpu)
		x.LoadUsed[cpu] = poolCore
		return cpu, true
	}
	return -1, false
}

// getFreeCorePhys - returns phys, logical, true (is_found)
func (x *PoolNodeInfo) getFreeCorePhys(physCore int) (int, int, bool) {
	poolCore, ok := x.LoadUsed[physCore]
	if ok {
		l0 := len(poolCore.Available)
		l1 := len(poolCore.Used)
		if l1 < l0 {
			core := poolCore.Available[l1]
			poolCore.Used = append(poolCore.Used, core)
			return physCore, core, true
		}
	}

	for physCore0, poolCore0 := range x.LoadFree {
		delete(x.LoadFree, physCore0)
		x.LoadUsed[physCore0] = poolCore0
		core := poolCore0.Available[0]
		poolCore0.Used = append(poolCore0.Used, core)
		return physCore0, core, true
	}

	return -1, -1, false
}

// freeCore - change core state from used to free
func (x *PoolNodeInfo) freeCore(core int) (int, bool) {
	if x.Config.LoadType == config.Phys {
		return x.freeCorePhys(core)
	}
	return x.freeCoreLogical(core)
}

func (x *PoolNodeInfo) freeCoreLogical(core int) (int, bool) {
	poolCore, ok := x.LoadUsed[core]
	if !ok {
		// core not exits in this numa!
		// or not used
		return 0, false
	}

	delete(x.LoadUsed, core)
	x.LoadFree[core] = poolCore
	return 1, true
}

func (x *PoolNodeInfo) freeCorePhys(core int) (int, bool) {
	for physCore, poolCore := range x.LoadUsed {
		for _, cpu := range poolCore.Used {
			if cpu == core {
				delete(x.LoadUsed, physCore)
				x.LoadFree[physCore] = poolCore
				count := len(poolCore.Used)
				poolCore.Used = poolCore.Used[0:0]
				return count, true
			}
		}
	}
	return 0, false
}

func mapIntToSlice(map0 map[int]*PoolCore) []int {
	l0 := len(map0)
	var sl0 []int
	if l0 > 0 {
		sl0 = make([]int, 0, l0)
		for key, _ := range map0 {
			sl0 = append(sl0, key)
		}
		sort.Slice(sl0, func(i, j int) bool {
			return sl0[i] < sl0[j]
		})
	}
	return sl0
}

func (x *PoolNodeInfo) getLoadUsedSlice() []*http_api.PoolCore {
	return mapToPoolCoresList(x.LoadUsed)
}

func (x *PoolNodeInfo) getLoadFreeSlice() []*http_api.PoolCore {
	return mapToPoolCoresList(x.LoadFree)
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

	sort.Slice(cpus, func(i, j int) bool {
		return cpus[i] < cpus[j]
	})
	return cpus
}

// isLoadCpuCountValid - checks is cpu count valid in mask
// of thread on load
func isLoadCpuCountValid(algo *config.PinCoresAlgo, count int) bool {
	return count == algo.Selected || count == algo.NotSelected
}
