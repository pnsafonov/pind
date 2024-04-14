package pkg

import (
	"github.com/pnsafonov/pind/pkg/config"
	"github.com/pnsafonov/pind/pkg/numa"
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
	LoadFree map[int]byte
	LoadUsed map[int]*PoolCore
}

// PoolCore - information about core pinned or free
type PoolCore struct {
	Evict bool     // not loaded, is ready for evict
	Proc  *PinProc // core is pinned to process
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
			LoadUsed: make(map[int]*PoolCore),
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

// getNumaNodeForLoadAssign - search for numa node changing starting node
// for config.EvictImmediately
func (x *Pool) getNumaNodeForLoadAssign(requiredCount int) (*PoolNodeInfo, bool) {
	evictMode := x.Config.Mode
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
		//freeCount := len(node.LoadFree)
		freeCount := node.getFreeCount(evictMode)
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
			count += thread.Cpus.AssignRequiredCores1(ctx, node, noSelected, &proc.NotSelected, proc)
			continue
		}
		if thread.Selected == ThreadSelectionYes {
			count += thread.Cpus.AssignRequiredCores0(ctx, node, selected, proc)
			continue
		}
		log.Warningf("PoolNodeInfo assignCores, execution must not be here!")
	}
	return count
}

// getFreeCore - returns free core
func (x *PoolNodeInfo) getFreeCore(proc *PinProc) (int, bool) {
	for cpu, _ := range x.LoadFree {
		delete(x.LoadFree, cpu)
		x.LoadUsed[cpu] = &PoolCore{
			Evict: false,
			Proc:  proc,
		}
		return cpu, true
	}
	return -1, false
}

// freeCore - change core state from used to free
func (x *PoolNodeInfo) freeCore(core int) bool {
	_, ok := x.LoadUsed[core]
	if !ok {
		// core not exits in this numa!
		// or not used
		return false
	}

	delete(x.LoadUsed, core)
	x.LoadFree[core] = 1
	return true
}

// markCoreForEvict - mark core for delayed evict
func (x *PoolNodeInfo) markCoreForEvict(core int, proc *PinProc) bool {
	poolCore, ok := x.LoadUsed[core]
	if !ok {
		return false
	}

	poolCore.Evict = true
	poolCore.Proc = proc
	return true
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

func mapIntToSlice0(map0 map[int]byte) []int {
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

func (x *PoolNodeInfo) getLoadUsedSlice() []int {
	return mapIntToSlice(x.LoadUsed)
}

func (x *PoolNodeInfo) getLoadFreeSlice() []int {
	return mapIntToSlice0(x.LoadFree)
}

// getFreeCount - get free cores count
func (x *PoolNodeInfo) getFreeCount(mode config.EvictMode) int {
	count0 := len(x.LoadFree)
	if mode == config.EvictImmediately {
		return count0
	}

	// config.EvictDelayed
	count1 := 0
	for _, pinCpu := range x.LoadUsed {
		if pinCpu.Evict {
			count1++
		}
	}
	return count0 + count1
}

func (x *PoolNodeInfo) getProcForEvict() (*PinProc, bool) {
	for _, pinCpu := range x.LoadUsed {
		if pinCpu.Evict {
			if pinCpu.Proc != nil {
				log.Errorf("getProcForEvict, pinCpu.Proc is nil")
				continue
			}
			return pinCpu.Proc, true
		}
	}
	return nil, false
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
