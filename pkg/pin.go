package pkg

import (
	"fmt"
	"github.com/prometheus/procfs"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"pind/pkg/numa"
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
	Procs map[int]*PinProc // pid -> PinProc, add/remove/change processes
	Used  map[int]*PinProc // cpu -> PinProc, add/remove/change used cpu
	Idle  PinCpus          // mask for idle threads
}

type PinProc struct {
	Proc     procfs.Proc // readonly
	ProcInfo *ProcInfo   // pointer changes every cycle

	Threads     map[int]*PinThread // pid -> PinThread, add/remove/update
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

func NewIdlePinCpu(ctx *Context) PinCpus {
	idle := ctx.Config.Service.Pool.Idle.Values
	l0 := len(idle)

	cpuSet := numa.CpusToMask(idle)
	pinCpus := PinCpus{
		Cpus:   make([]int, l0),
		CpuSet: cpuSet,
	}
	copy(pinCpus.Cpus, idle)

	return pinCpus
}

func NewPinState() PinState {
	state := PinState{
		Procs: make(map[int]*PinProc),
		Used:  make(map[int]*PinProc),
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
			freeThreadCpus(state, thread)
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

// freeThreadCpus - free cpu's list used pinned to thread
func freeThreadCpus(state *PinState, thread *PinThread) {
	l0 := len(thread.Cpus.Cpus)
	for i := 0; i < l0; i++ {
		cpu := thread.Cpus.Cpus[i]
		log.Debugf("freeThreadCpus pid = %d, comm = %s, cpu = %d, used = %v", thread.ThreadInfo.Stat.PID, thread.ThreadInfo.Stat.Comm, cpu, state.Used)
		delete(state.Used, cpu)
	}
	thread.Cpus.Zero()
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
			freeThreadCpus(state, thread)

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
		if threadInfo.Selected == selection {
			return true
		}
	}
	return false
}

func (x *PinProc) ContainsNotSelectedThread() bool {
	return x.ContainsThread(ThreadSelectionNo)
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
				freeThreadCpus(state, thread)
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
	}

	return err
}

func (x *PinState) PinLoad(ctx *Context) error {
	var err error
	state := x
	patterns := ctx.Config.Service.Selection.Patterns
	algo := ctx.Config.Service.PinCoresAlgo

	// update selection
	for _, procInfo := range state.Procs {
		if !procInfo.ProcInfo.load {
			continue
		}
		for _, threadInfo := range procInfo.Threads {
			if threadInfo.Selected == ThreadSelectionYes || threadInfo.Selected == ThreadSelectionNo {
				continue
			}
			selection := getThreadSelection(threadInfo.ThreadInfo, patterns)
			threadInfo.Selected = selection
		}
	}

	// assign masks
	for _, procInfo := range state.Procs {
		if !procInfo.ProcInfo.load {
			continue
		}
		// we must check, that at least 1 not selected thread exists
		containsNotSelectedThread := procInfo.ContainsNotSelectedThread()
		if containsNotSelectedThread {
			// we must init mask for not selected threads
			if !procInfo.NotSelected.IsInited(algo.NotSelected) {
				err0 := procInfo.NotSelected.AssignCores(ctx, algo.NotSelected, procInfo)
				if err0 != nil {
					err = err0
				}
			}
		}

		for _, threadInfo := range procInfo.Threads {
			if threadInfo.Selected == ThreadSelectionNo {
				if !threadInfo.Cpus.IsInited(algo.NotSelected) {
					// use same mask for not selected threads
					threadInfo.Cpus = procInfo.NotSelected
				}
				continue
			}
			if threadInfo.Selected == ThreadSelectionYes {
				if !threadInfo.Cpus.IsInited(algo.Selected) {
					// use different cores for selected threads
					err0 := threadInfo.Cpus.AssignCores(ctx, algo.Selected, procInfo)
					if err0 != nil {
						err = err0
					}
				}
				continue
			}
			log.Errorf("PinLoad, execution must not be here!")
		}
	}

	// pin cores
	for _, procInfo := range state.Procs {
		if !procInfo.ProcInfo.load {
			continue
		}
		for _, threadInfo := range procInfo.Threads {
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

func (x *PinCpus) AssignCores(ctx *Context, count int, procInfo *PinProc) error {
	err := ErrNoFreeCores
	load := ctx.Config.Service.Pool.Load
	used := ctx.state.Used

	l0 := len(load.Values)
	for i := 0; i < l0; i++ {
		cpu := load.Values[i]
		_, ok := used[cpu]
		if ok {
			continue
		}

		used[cpu] = procInfo
		x.Cpus = append(x.Cpus, cpu)
		log.Debugf("AssignCores pid = %d, comm = %s, cpu = %d, used = %v", procInfo.ProcInfo.Stat.PID, procInfo.ProcInfo.Stat.Comm, cpu, used)
		if len(x.Cpus) >= count {
			err = nil
			break
		}
	}

	x.CpuSet = numa.CpusToMask(x.Cpus)
	return err
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
