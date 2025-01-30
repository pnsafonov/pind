package pkg

import (
	"fmt"
	"sort"

	"github.com/pnsafonov/pind/pkg/config"
	"github.com/pnsafonov/pind/pkg/numa"
	"github.com/pnsafonov/pind/pkg/utils/core_utils"
	"github.com/pnsafonov/pind/pkg/utils/math_utils"
	"github.com/prometheus/procfs"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type ThreadSelection int

const (
	ThreadSelectionUnknown ThreadSelection = 0 // unknown
	ThreadSelectionYes     ThreadSelection = 1 // selected
	ThreadSelectionNo      ThreadSelection = 2 // not selected
)

var (
	ErrNoFreeCores               = fmt.Errorf("no_free_cores_left")
	ErrSchedSetAffinityEmptyMask = fmt.Errorf("sched_setaffinity_empty_mask")
)

type PinState struct {
	Procs  map[int]*PinProc // pid -> PinProc, add/remove/change processes
	Idle   PinCpus          // mask for idle threads
	Errors *Errors
}

type PinProc struct {
	Proc     procfs.Proc // readonly
	ProcInfo *ProcInfo   // pointer changes every cycle

	Threads     map[int]*PinThread // pid -> PinThread, add/remove/update
	Node        *PoolNodeInfo      // process pinned to numa node
	NotSelected PinCpus            // cores for not selected threads
}

type PinThread struct {
	Thread     procfs.Proc // readonly
	ThreadInfo *ThreadInfo // readonly, actual information

	Cpus     PinCpus         // what we want for selected thread
	Selected ThreadSelection // is thread selected
}

type PinCpus struct {
	Cpus   []int       // must be assigned, same as CpuSet
	CpuSet unix.CPUSet // must be assigned, same as Cpus
}

type Errors struct {
	CalcProcsCPU         error
	CalcCoresCPU         error
	PinNotInFilterToIdle error
	StatePinIdle         error
	StatePinLoad         error
	IdleOverwork         error
	RequiredCPU          RequiredCPU
}

type RequiredCPU struct {
	Total      int
	PerProcess []int
}

func NewIdlePinCpu(ctx *Context) PinCpus {
	idle := ctx.Config.Service.Pool.Idle.Values

	cpus := core_utils.CopyIntSlice(idle)
	cpuSet := numa.CpusToMask(idle)
	pinCpus := PinCpus{
		Cpus:   cpus,
		CpuSet: cpuSet,
	}

	return pinCpus
}

func NewPinState() PinState {
	state := PinState{
		Procs:  make(map[int]*PinProc),
		Errors: &Errors{},
	}
	return state
}

func NewPinProc(proc *ProcInfo) *PinProc {
	pinProc := &PinProc{
		Proc:     proc.Proc,
		ProcInfo: proc,
		Threads:  make(map[int]*PinThread),
	}

	l0 := len(proc.Threads)
	for i := 0; i < l0; i++ {
		thread := proc.Threads[i]
		pinThread := NewPinThread(thread)
		pinProc.Threads[pinThread.Thread.PID] = pinThread
	}

	return pinProc
}

func NewPinThread(thread *ThreadInfo) *PinThread {
	pinThread := &PinThread{
		Thread:     thread.Thread,
		ThreadInfo: thread,
	}
	return pinThread
}

func getProcByPID(procs []*ProcInfo, pid int) (*ProcInfo, bool) {
	l0 := len(procs)
	for i := 0; i < l0; i++ {
		proc := procs[i]
		if proc.Proc.PID == pid {
			return proc, true
		}
	}
	return nil, false
}

func (x *PinState) UpdateProcs(procs []*ProcInfo) {
	state := x

	// remove
	for pid, proc := range state.Procs {
		_, ok0 := getProcByPID(procs, pid)
		if ok0 {
			// update actions implemented below
			continue
		}
		// free cpu on process end
		for _, thread := range proc.Threads {
			freeThreadCpus0(proc, thread)
		}

		// process is finished
		delete(state.Procs, pid)
	}

	// add
	l0 := len(procs)
	for i := 0; i < l0; i++ {
		proc := procs[i]

		pinProc, ok := state.Procs[proc.Proc.PID]
		if !ok {
			pinProc = NewPinProc(proc)
			state.Procs[proc.Proc.PID] = pinProc
		} else {
			// update actions
			pinProc.UpdateProc(proc, state)
		}
	}
}

func freeThreadCpus0(proc *PinProc, thread *PinThread) int {
	node := proc.Node
	if node == nil {
		// numa node is not assigned
		thread.Cpus.Zero()
		return 0
	}

	count := 0
	l0 := len(thread.Cpus.Cpus)
	for i := 0; i < l0; i++ {
		cpu := thread.Cpus.Cpus[i]
		count0, result := node.freeCore(cpu)
		if result {
			count += count0
		}
		log.Debugf("freeThreadCpus pid = %d, comm = %s, cpu = %d, result = %v", thread.ThreadInfo.Stat.PID, thread.ThreadInfo.Stat.Comm, cpu, result)
	}
	thread.Cpus.Zero()
	return count
}

func getThreadByPID(threads []*ThreadInfo, pid int) (*ThreadInfo, bool) {
	l0 := len(threads)
	for i := 0; i < l0; i++ {
		thread := threads[i]
		if thread.Thread.PID == pid {
			return thread, true
		}
	}
	return nil, false
}

func (x *PinProc) UpdateProc(proc *ProcInfo, state *PinState) {
	// remove thread
	for pid, thread := range x.Threads {
		_, ok := getThreadByPID(proc.Threads, pid)
		if ok {
			// update actions implemented below
			continue
		}

		needFree := isThreadCanBeFreed(thread, x)
		if needFree {
			// free cpu on thread end
			freeThreadCpus0(x, thread)

			if thread.Selected == ThreadSelectionNo {
				// after free not selected threads, reset cpus cores pattern
				x.NotSelected.Zero()
			}
		}
		// thread is finished
		delete(x.Threads, pid)
	}

	// add thread
	l0 := len(proc.Threads)
	for i := 0; i < l0; i++ {
		thread := proc.Threads[i]

		pinThread, ok := x.Threads[thread.Thread.PID]
		if !ok {
			pinThread = NewPinThread(thread)
			x.Threads[thread.Thread.PID] = pinThread
		} else {
			// update actions
			pinThread.UpdateThread(thread)
		}
	}

	x.ProcInfo = proc
}

// isThreadCanBeFreed - checks what that can be deleted
// only last deleting not selected thread must free cpu cores
func isThreadCanBeFreed(thread *PinThread, pinProc *PinProc) bool {
	if thread.Selected != ThreadSelectionNo {
		return true
	}
	// not selected thread must be checked
	count := getThreadsCount0(pinProc.Threads, ThreadSelectionNo)
	return count <= 1
}

// getThreadsCount0 - get threads count with selection
func getThreadsCount0(threads map[int]*PinThread, selection ThreadSelection) int {
	count := 0
	for _, thread := range threads {
		if thread.Selected == selection {
			count++
		}
	}
	return count
}

func (x *PinProc) ContainsThread(selection ThreadSelection) bool {
	procInfo := x
	for _, threadInfo := range procInfo.Threads {
		if threadInfo.ThreadInfo.Ignored {
			continue
		}
		if threadInfo.Selected == selection {
			return true
		}
	}
	return false
}

func (x *PinProc) ContainsNotSelectedThread() bool {
	return x.ContainsThread(ThreadSelectionNo)
}

// getRequiredCpuCount - returns phys count, logical count
func (x *PinProc) getRequiredCpuCount(ctx *Context) (int, int) {
	requiredNotSelected := ctx.Config.Service.PinCoresAlgo.NotSelected
	requiredSelected := ctx.Config.Service.PinCoresAlgo.Selected

	notSelectedPhys := 0
	notSelected := 0
	if x.ContainsNotSelectedThread() {
		notSelected = x.NotSelected.getRequiredCpuCount(requiredNotSelected)
		notSelectedPhys = math_utils.IntDivide2Ceil(notSelected)
	}

	selectedPhys := 0
	selected := 0
	for _, thread := range x.Threads {
		if thread.ThreadInfo.Ignored {
			continue
		}
		if thread.Selected != ThreadSelectionYes {
			continue
		}
		selected0 := thread.Cpus.getRequiredCpuCount(requiredSelected)
		selectedPhys0 := math_utils.IntDivide2Ceil(selected0)
		selected += selected0
		selectedPhys += selectedPhys0
	}

	return notSelectedPhys + selectedPhys, notSelected + selected
}

// UpdateThread - set actual information abount thread
func (x *PinThread) UpdateThread(thread *ThreadInfo) {
	x.ThreadInfo = thread
}

func (x *PinState) PinIdle() error {
	var err error
	state := x

	for _, procInfo := range state.Procs {
		if procInfo.ProcInfo.load {
			continue
		}

		for _, thread := range procInfo.Threads {
			if thread.Cpus.IsAnyInited() {
				freeThreadCpus0(procInfo, thread)
			}

			if thread.ThreadInfo.Ignored {
				continue
			}

			if isMasksEqual(thread.ThreadInfo.CpuSet, state.Idle.CpuSet) {
				continue
			}

			err0 := schedSetAffinity(thread.ThreadInfo.Stat, &state.Idle.CpuSet)
			if err0 != nil {
				err = err0
			}
		}

		if procInfo.NotSelected.IsAnyInited() {
			// is not used in idle
			procInfo.NotSelected.Zero()
		}

		// reset node when idle
		// error: core is not freed in load pool
		//procInfo.Node = nil
	}

	return err
}

func (x *PinState) PinLoad(ctx *Context) error {
	var err error
	state := x
	errs := state.Errors
	patterns := ctx.Config.Service.Selection.Patterns
	//algo := ctx.Config.Service.PinCoresAlgo
	pool := ctx.pool

	// update selection
	for _, procInfo := range state.Procs {
		if !procInfo.ProcInfo.load {
			continue
		}
		for _, threadInfo := range procInfo.Threads {
			if threadInfo.ThreadInfo.Ignored {
				continue
			}
			if threadInfo.Selected == ThreadSelectionYes || threadInfo.Selected == ThreadSelectionNo {
				continue
			}
			selection := getThreadSelection(threadInfo.ThreadInfo, patterns)
			threadInfo.Selected = selection
		}
	}

	// assign numa nodes, assign cpu cores
	for _, procInfo := range state.Procs {
		if !procInfo.ProcInfo.load {
			continue
		}

		cpuCountPhys, cpuCount := procInfo.getRequiredCpuCount(ctx)
		if cpuCount <= 0 {
			continue
		}

		node := procInfo.Node
		if node == nil {
			node0, ok := pool.getNumaNodeForLoadAssign(cpuCountPhys, cpuCount)
			if !ok {
				if ctx.Config.Service.Pool.PinMode == config.PinModeDelayed {
					// evict delayed vms to idle cores
					err = state.PinIdle()
					if err != nil {
						log.Errorf("PinLoad, PinIdle 0 err = %v", err)
						continue
					}
					// second attempt after clean delayed
					node0, ok = pool.getNumaNodeForLoadAssign(cpuCountPhys, cpuCount)
				}
				if !ok {
					vmName, _ := parseVmName(procInfo.ProcInfo.Cmd)
					err = fmt.Errorf("PinState, PinLoad pool.getNumaNodeForLoadAssign failed for cpuCountPhys = %d, cpuCount = %d, vmName = %s", cpuCountPhys, cpuCount, vmName)
					errs.RequiredCPU.addCpuCount(cpuCount)
					continue
				}
			}
			procInfo.Node = node0
			node = node0
		}

		assignedCount := node.assignCores(ctx, procInfo)
		if assignedCount != cpuCount {
			if ctx.Config.Service.Pool.PinMode == config.PinModeDelayed {
				// evict delayed vms to idle cores
				err = state.PinIdle()
				if err != nil {
					log.Errorf("PinLoad, PinIdle 1 err = %v", err)
					continue
				}
				// second attempt to assign cores
				assignedCount = node.assignCores(ctx, procInfo)
			}

			if assignedCount != cpuCount {
				vmName, _ := parseVmName(procInfo.ProcInfo.Cmd)
				nodeState := node.StateString()
				log.Warningf("PinState PinLoad, node.assignCores failed, assignedCount = %d need cpuCount = %d, vmName = %s, %s", assignedCount, cpuCount, vmName, nodeState)
			}
		}
	}

	// pin cores
	for _, procInfo := range state.Procs {
		if !procInfo.ProcInfo.load {
			continue
		}
		for _, threadInfo := range procInfo.Threads {
			if threadInfo.ThreadInfo.Ignored {
				continue
			}

			if !threadInfo.Cpus.IsAnyInited() {
				// no free cpu's cores left
				continue
			}

			if isMasksEqual(threadInfo.ThreadInfo.CpuSet, threadInfo.Cpus.CpuSet) {
				// actual mask is set
				continue
			}

			err1 := schedSetAffinity(threadInfo.ThreadInfo.Stat, &threadInfo.Cpus.CpuSet)
			if err1 != nil {
				err = err1
			}
		}
	}

	return err
}

func schedSetAffinity(procStat procfs.ProcStat, set *unix.CPUSet) error {
	pid := procStat.PID
	var cpus []int

	count0 := set.Count()
	if count0 == 0 {
		log.Errorf("schedSetAffinity, tries to set zero mask, pid = %d, comm = %s", pid, procStat.Comm)
		return ErrSchedSetAffinityEmptyMask
	}

	isDebugEnabled := log.IsLevelEnabled(log.DebugLevel)
	isErrorEnabled := log.IsLevelEnabled(log.ErrorLevel)
	if isDebugEnabled || isErrorEnabled {
		cpus = MaskToArray(set)
	}

	if isDebugEnabled {
		log.Debugf("schedSetaffinity pid = %d, comm = %s, cpus = %v", pid, procStat.Comm, cpus)
	}
	err := unix.SchedSetaffinity(pid, set)
	if err != nil && isErrorEnabled {
		log.Errorf("schedSetaffinity err = %v, pid = %d, comm = %s, cpus = %v", err, pid, procStat.Comm, cpus)
	}
	return err
}

func (x *PinCpus) Zero() {
	x.Cpus = x.Cpus[:0]
	x.CpuSet = unix.CPUSet{}
}

func (x *PinCpus) IsAnyInited() bool {
	l0 := len(x.Cpus)
	return l0 > 0
}

func (x *PinCpus) IsInited(count int) bool {
	l0 := len(x.Cpus)
	return l0 >= count
}

func (x *PinCpus) getRequiredCpuCount(requiredCount int) int {
	l0 := len(x.Cpus)
	if l0 >= requiredCount {
		return 0
	}
	return requiredCount - l0
}

func (x *PinCpus) AssignRequiredCores0(node *PoolNodeInfo, count int) int {
	requiredCount := x.getRequiredCpuCount(count)
	if requiredCount <= 0 {
		return 0
	}
	assignedCount := 0
	physCpu := -1
	cpu := -1
	ok := false
	for ; assignedCount < requiredCount; assignedCount++ {
		physCpu, cpu, ok = node.getFreeCore(physCpu)
		if !ok {
			break
		}
		x.Cpus = append(x.Cpus, cpu)
	}
	//if assignedCount != requiredCount {
	//	log.Warningf("PinCpus AssignRequiredCores0, assignedCount = %d, requiredCount = %d", assignedCount, requiredCount)
	//}
	if assignedCount > 0 {
		x.CpuSet = numa.CpusToMask(x.Cpus)
		sort.Slice(x.Cpus, func(i, j int) bool {
			return x.Cpus[i] < x.Cpus[j]
		})
	}
	return assignedCount
}

// AssignRequiredCores1 - use shared cores, for not selected threads
func (x *PinCpus) AssignRequiredCores1(node *PoolNodeInfo, count int, shared *PinCpus) int {
	count0 := shared.AssignRequiredCores0(node, count)

	if !isMasksEqual(x.CpuSet, shared.CpuSet) {
		x.assignAsCopy(shared)
	}

	return count0
}

func (x *PinCpus) assignAsCopy(right *PinCpus) {
	x.Cpus = core_utils.CopyIntSlice(right.Cpus)
	x.CpuSet = right.CpuSet
}

func (x *PinCpus) getCpusCopy() []int {
	return core_utils.CopyIntSlice(x.Cpus)
}

func (x *PinCpus) getCpusCopy0() *[]int {
	result := new([]int)
	*result = x.getCpusCopy()
	return result
}

func pinNotInFilterToIdle(ctx *Context) error {
	var err error
	state := ctx.state

	l0 := len(ctx.lastNotInFilter)
	for i := 0; i < l0; i++ {
		proc := ctx.lastNotInFilter[i]

		l1 := len(proc.Threads)
		for j := 0; j < l1; j++ {
			thread := proc.Threads[j]

			if thread.Ignored {
				continue
			}

			if isMasksEqual(state.Idle.CpuSet, thread.CpuSet) {
				continue
			}

			err0 := schedSetAffinity(thread.Stat, &state.Idle.CpuSet)
			if err0 != nil {
				err = err0
			}
		}
	}
	return err
}

func (x *Errors) clear() {
	x.CalcProcsCPU = nil
	x.CalcCoresCPU = nil
	x.PinNotInFilterToIdle = nil
	x.StatePinIdle = nil
	x.StatePinLoad = nil
	x.IdleOverwork = nil
	x.RequiredCPU.Total = 0
	x.RequiredCPU.PerProcess = x.RequiredCPU.PerProcess[:0]
}

func (x *RequiredCPU) addCpuCount(count int) {
	x.Total += count
	x.PerProcess = append(x.PerProcess, count)
}
