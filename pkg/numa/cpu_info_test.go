package numa

import (
	"github.com/prometheus/procfs/sysfs"
	"testing"
)

func TestCpuInfo(t *testing.T) {
	cpuInfos, err := procfs0.CPUInfo()
	if err != nil {
		t.FailNow()
	}
	_ = cpuInfos
}

func TestCpuTopology(t *testing.T) {
	fs0, err := sysfs.NewDefaultFS()
	if err != nil {
		t.FailNow()
	}

	cpus, err := fs0.CPUs()
	if err != nil {
		t.FailNow()
	}

	l0 := len(cpus)
	for i := 0; i < l0; i++ {
		cpu0 := cpus[i]
		topology, err := cpu0.Topology()
		if err != nil {
			t.FailNow()
		}
		_ = topology

		number := cpu0.Number()
		_ = number

		throttle, err := cpu0.ThermalThrottle()
		if err != nil {
			//t.FailNow()
			t.Logf("cpu0.ThermalThrottle err = %v\n", err)
		}
		_ = throttle
	}
}
