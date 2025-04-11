package pkg

import (
	"github.com/pnsafonov/pind/pkg/config"
	"github.com/prometheus/procfs"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"sort"
	"strings"
	"time"
)

type ProcInfo struct {
	Proc    procfs.Proc
	Stat    procfs.ProcStat
	Cmd     []string
	Threads []*ThreadInfo

	VmName string

	time time.Time
	cpu0 float64
	load bool
}

// ThreadInfo - current (actual) information about thread
// filled once, and not modified
type ThreadInfo struct {
	Thread  procfs.Proc
	Stat    procfs.ProcStat
	CpuSet  unix.CPUSet
	Ignored bool // ignore thread: can't sched_getaffinity with any cpu
}

func NewProcInfo(proc procfs.Proc, procStat procfs.ProcStat, cmd0 []string) *ProcInfo {
	procInfo := &ProcInfo{
		Proc: proc,
		Stat: procStat,
		Cmd:  cmd0,
	}
	return procInfo
}

func filterProcsInfo0(filters []*config.ProcFilter, filtersAlwaysIdle []*config.ProcFilter, ignore *config.Ignore) ([]*ProcInfo, []*ProcInfo, error) {
	procs, err := procfs.AllProcs()
	if err != nil {
		log.Errorf("filterProcsInfo0 procfs.AllProcs err = %v", err)
		return nil, nil, err
	}

	result := make([]*ProcInfo, 0, 16)
	procsAlwaysIdle := make([]*ProcInfo, 0, 16)
	l0 := len(procs)
	for i := 0; i < l0; i++ {
		proc := procs[i]
		procStat, err0 := proc.Stat()
		if err0 != nil {
			continue
		}

		cmd0, err0 := proc.CmdLine()
		if err0 != nil {
			continue
		}

		if filterProc(filtersAlwaysIdle, procStat.Comm, cmd0) {
			// always idle processes
			procInfo := NewProcInfo(proc, procStat, cmd0)
			err = procInfoAddThreads(proc, procInfo, nil)
			if err == nil {
				procsAlwaysIdle = append(procsAlwaysIdle, procInfo)
			}
			continue
		}

		if filterProc(filters, procStat.Comm, cmd0) {
			// process to pin
			procInfo := NewProcInfo(proc, procStat, cmd0)
			err = procInfoAddThreads(proc, procInfo, ignore)
			if err == nil {
				result = append(result, procInfo)
			}
			//continue
		}
	}

	// sort by id
	sort.Slice(result, func(i, j int) bool {
		return result[i].Proc.PID < result[j].Proc.PID
	})

	// init virtual machine name
	for _, procInfo := range result {
		vmName, ok := parseVmName(procInfo.Cmd)
		if ok {
			procInfo.VmName = vmName
		}
	}

	return result, procsAlwaysIdle, nil
}

func procInfoAddThreads(proc procfs.Proc, procInfo *ProcInfo, ignore *config.Ignore) error {
	threads, err0 := procfs.AllThreads(proc.PID)
	if err0 != nil {
		return err0
	}
	l1 := len(threads)
	for j := 0; j < l1; j++ {
		thread := threads[j]
		threadStat, err1 := thread.Stat()
		if err1 != nil {
			return err1
		}

		ignored := isIgnored(threadStat.Comm, ignore)

		threadInfo := &ThreadInfo{
			Thread:  thread,
			Stat:    threadStat,
			Ignored: ignored,
		}
		// sched_getaffinity
		err1 = unix.SchedGetaffinity(threadStat.PID, &threadInfo.CpuSet)
		if err1 != nil {
			return err1
		}

		procInfo.Threads = append(procInfo.Threads, threadInfo)
	}
	return nil
}

// filterProc
// true  -     matched by filter
// false - not matched by filter
func filterProc(filters []*config.ProcFilter, comm string, cmd0 []string) bool {
	l0 := len(filters)
	for i := 0; i < l0; i++ {
		filter := filters[i]
		if filterProc0(filter, comm, cmd0) {
			return true
		}
	}
	return false
}

func filterProc0(filter *config.ProcFilter, comm string, cmd0 []string) bool {
	l0 := len(filter.Patterns)
	for i := 0; i < l0; i++ {
		pattern := filter.Patterns[i]
		if strings.Contains(comm, pattern) {
			continue
		}
		if arrayContainsPattern(cmd0, pattern) {
			continue
		}
		return false
	}
	return true
}

func arrayContainsPattern(arr []string, pattern string) bool {
	l0 := len(arr)
	for i := 0; i < l0; i++ {
		str := arr[i]
		if strings.Contains(str, pattern) {
			return true
		}
	}
	return false
}

// isProcInfoSame
// true  - the same process
// false - different processes
func isProcInfoSame(pi0 *ProcInfo, pi1 *ProcInfo) bool {
	if pi0 == nil || pi1 == nil {
		return false
	}
	if pi0.Proc.PID != pi1.Proc.PID {
		return false
	}
	if pi0.Stat.Starttime != pi1.Stat.Starttime {
		return false
	}
	return true
}

func getSameProc(procs0 []*ProcInfo, proc1 *ProcInfo) (*ProcInfo, bool) {
	l0 := len(procs0)
	for i := 0; i < l0; i++ {
		proc0 := procs0[i]
		if isProcInfoSame(proc0, proc1) {
			return proc0, true
		}
	}
	return nil, false
}

// isThreadSelected
// true  - thread selected by pattern
// false - thread not selected
func isThreadSelected(thead *ThreadInfo, patterns []string) bool {
	l1 := len(patterns)
	for j := 0; j < l1; j++ {
		pattern := patterns[j]
		if strings.Contains(thead.Stat.Comm, pattern) {
			return true
		}
	}
	return false
}

func getThreadSelection(thead *ThreadInfo, patterns []string) ThreadSelection {
	isSelected := isThreadSelected(thead, patterns)
	if isSelected {
		return ThreadSelectionYes
	}
	return ThreadSelectionNo
}

// filterProcInfo - returns (inFilter, notInFilter)
func filterProcInfo(procs []*ProcInfo, filters []*config.ProcFilter) ([]*ProcInfo, []*ProcInfo) {
	l0 := len(procs)

	inFilter := make([]*ProcInfo, 0, l0)
	notInFilter := make([]*ProcInfo, 0, l0)

	for i := 0; i < l0; i++ {
		proc := procs[i]

		if filterProc(filters, proc.Stat.Comm, proc.Cmd) {
			inFilter = append(inFilter, proc)
		} else {
			notInFilter = append(notInFilter, proc)
		}
	}

	return inFilter, notInFilter
}

// isIgnored - is thread ignored, contains any of ignore pattern
func isIgnored(comm string, ignore *config.Ignore) bool {
	if ignore == nil {
		return false
	}
	l0 := len(ignore.Patterns)
	for i := 0; i < l0; i++ {
		pattern := ignore.Patterns[i]
		if strings.Contains(comm, pattern) {
			return true
		}
	}
	return false
}

func parseVmName(cmd []string) (string, bool) {
	l0 := len(cmd)
	for i := 0; i < l0; i++ {
		arg0 := strings.TrimSpace(cmd[i])
		if arg0 == "-name" {
			i1 := i + 1
			if i1 < l0 {
				arg1 := cmd[i1]
				spli1 := strings.Split(arg1, ",")
				l1 := len(spli1)

				// mz-pgpro-8796-ent-load,debug-threads=on
				for j := 0; j < l1; j++ {
					str0 := strings.TrimSpace(spli1[j])
					if !strings.Contains(str0, "=") {
						return str0, true
					}
				}

				// guest=deb-3,debug-threads=on
				for j := 0; j < l1; j++ {
					str1 := strings.TrimSpace(spli1[j])
					if strings.Contains(str1, "=") {
						split2 := strings.Split(str1, "=")
						l2 := len(split2)
						if l2 >= 2 {
							key0 := strings.TrimSpace(split2[0])
							val0 := strings.TrimSpace(split2[1])
							if key0 == "guest" {
								return val0, true
							}
						}
					}
				}
			}
		}
	}
	return "", false
}
