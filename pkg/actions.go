package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"pind/pkg/config"
)

func PrintProcs1(patterns []string) error {
	filters := []*config.ProcFilter{
		&config.ProcFilter{
			Patterns: patterns,
		},
	}
	return PrintProcs0(filters)
}

func PrintProcs2() error {
	filters := config.NewDefaultFilters()
	return PrintProcs0(filters)
}

func PrintProcs0(filters []*config.ProcFilter) error {
	procs, err := filterProcsInfo0(filters)
	if err != nil {
		log.Errorf("PrintProcs0 filterProcsInfo0 err = %v", err)
		return err
	}

	l0 := len(procs)
	for i := 0; i < l0; i++ {
		proc := procs[i]
		fmt.Printf("% 6d %s %v\n", proc.Proc.PID, proc.Stat.Comm, proc.Cmd)
		proc.Stat.CPUTime()
		l1 := len(proc.Threads)
		for j := 0; j < l1; j++ {
			thread := proc.Threads[j]

			fmt.Printf("% 10d %s\n", thread.Thread.PID, thread.Stat.Comm)
		}

		fmt.Printf("\n")
	}

	return nil
}