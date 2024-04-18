package numa

import "testing"

func TestCPUInfo(t *testing.T) {
	cpuInfos, err := procfs0.CPUInfo()
	if err != nil {
		t.FailNow()
	}
	_ = cpuInfos
}
