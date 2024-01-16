package pkg

import (
	"fmt"
	"github.com/prometheus/procfs"
	"golang.org/x/sys/unix"
	"time"
)

func GetProcs0() error {
	procs, err := procfs.AllProcs()
	if err != nil {
		return err
	}

	l0 := len(procs)
	for i := 0; i < l0; i++ {
		proc := procs[i]
		procStat, err0 := proc.Stat()
		if err0 != nil {
			continue
		}
		//fmt.Printf("%06d %s\n", proc.PID, procStat.Comm)
		fmt.Printf("% 6d %s\n", proc.PID, procStat.Comm)

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
			fmt.Printf("% 10d %s\n", threadStat.PID, threadStat.Comm)

			// sched_getaffinity
			var cpuset unix.CPUSet
			err1 = unix.SchedGetaffinity(threadStat.PID, &cpuset)
			if err1 != nil {
				continue
			}

			//unix.SchedSetaffinity()
		}

	}

	return nil
}

func DoTicker() {
	//ticker := time.NewTicker(500 * time.Millisecond)
	ticker := time.NewTicker(1000 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Before sleep", t)
				time.Sleep(3 * time.Second)
				fmt.Println("After Sleep", t)
			}
		}
	}()

	//time.Sleep(1600 * time.Millisecond)
	time.Sleep(30 * time.Second)
	ticker.Stop()
	done <- true
	fmt.Println("Ticker stopped")
}
