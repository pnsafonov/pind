package pkg

import (
	"github.com/prometheus/procfs"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"pind/pkg/config"
	"strings"
)

type ProcInfo struct {
	Proc    procfs.Proc
	Stat    procfs.ProcStat
	Cmd     []string
	Threads []*ThreadInfo
}

type ThreadInfo struct {
	Thread procfs.Proc
	Stat   procfs.ProcStat
	CpuSet unix.CPUSet
}

func filterProcsInfo0(filters []*config.ProcFilter) ([]*ProcInfo, error) {
	procs, err := procfs.AllProcs()
	if err != nil {
		log.Errorf("filterProcsInfo0 procfs.AllProcs err = %v", err)
		return nil, err
	}

	result := make([]*ProcInfo, 0, 16)
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

		if !filterProc(filters, procStat.Comm, cmd0) {
			continue
		}

		procInfo := &ProcInfo{
			Proc: proc,
			Stat: procStat,
			Cmd:  cmd0,
		}
		result = append(result, procInfo)

		threads, err0 := procfs.AllThreads(proc.PID)
		if err0 != nil {
			continue
		}
		l1 := len(threads)
		for j := 0; j < l1; j++ {
			thread := threads[j]
			threadStat, err1 := thread.Stat()
			if err1 != nil {
				continue
			}

			threadInfo := &ThreadInfo{
				Thread: thread,
				Stat:   threadStat,
			}
			// sched_getaffinity
			err1 = unix.SchedGetaffinity(threadStat.PID, &threadInfo.CpuSet)
			if err1 != nil {
				continue
			}

			procInfo.Threads = append(procInfo.Threads, threadInfo)
		}

	}

	return result, nil
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
