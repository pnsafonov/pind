package numa

import (
	"fmt"
	"github.com/prometheus/procfs/sysfs"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type CpuTopologyInfo struct {
	Cpu         sysfs.CPU
	Number      string
	CpuTopology *sysfs.CPUTopology // text
	Topology    *Topology
}

// Topology - is parsed sysfs.CPUTopology
type Topology struct {
	CoreID             int
	CoreSiblingsList   []int
	PhysicalPackageID  int
	ThreadSiblingsList []int
}

func PrintTopology() error {
	infos, err := GetCpuTopologyInfo()
	if err != nil {
		log.Errorf("PrintTopology, GetCpuTopologyInfo err = %v", err)
		return err
	}

	l0 := len(infos)
	for i := 0; i < l0; i++ {
		info := infos[i]
		tpg := info.CpuTopology
		fmt.Printf("num=%s, coreId=%s, core_siblings=%s; phys=%s, thread_siblings=%s\n",
			info.Number, tpg.CoreID, tpg.CoreSiblingsList, tpg.PhysicalPackageID,
			tpg.ThreadSiblingsList)
	}
	return nil
}

func GetCpuTopologyInfo() ([]*CpuTopologyInfo, error) {
	err := checkInitError()
	if err != nil {
		return nil, err
	}

	cpus0, err := sysfs0.CPUs()
	if err != nil {
		log.Errorf("PrintTopology, sysfs0.CPUs err = %v", err)
		return nil, err
	}

	l0 := len(cpus0)
	infos := make([]*CpuTopologyInfo, 0, l0)
	for i := 0; i < l0; i++ {
		cpu0 := cpus0[i]
		number := cpu0.Number()
		cpuTopology, err := cpu0.Topology()
		if err != nil {
			log.Errorf("PrintTopology, cpu0.Topology err = %v", err)
			return nil, err
		}

		topology, err := newTopology(cpuTopology)
		if err != nil {
			log.Errorf("PrintTopology, newTopology err = %v", err)
			return nil, err
		}

		ti := &CpuTopologyInfo{
			Number:      number,
			Cpu:         cpu0,
			CpuTopology: cpuTopology,
			Topology:    topology,
		}
		infos = append(infos, ti)
	}
	return infos, nil
}

// parseIntList - parse list like "1,7,8,16"
func parseIntList(val0 string) ([]int, error) {
	split0 := strings.Split(val0, ",")
	l0 := len(split0)
	result := make([]int, 0, l0)
	for i := 0; i < l0; i++ {
		str0 := strings.TrimSpace(split0[i])
		val, err := strconv.Atoi(str0)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}
	return result, nil
}

// parseIntList - parse list like "0-15"
func parseIntList0(val0 string) ([]int, error) {
	split0 := strings.Split(val0, "-")
	l0 := len(split0)
	if l0 != 2 {
		//log.Errorf("parseIntList0, invalid split count = %d", l0)
		//return nil, fmt.Errorf("invalid split count = %d", l0)
		return nil, nil
	}

	str0 := strings.TrimSpace(split0[0])
	from, err := strconv.Atoi(str0)
	if err != nil {
		return nil, err
	}

	str1 := strings.TrimSpace(split0[1])
	to, err := strconv.Atoi(str1)
	if err != nil {
		return nil, err
	}

	l1 := to - from
	l1 += 1
	if l1 < 0 {
		return nil, fmt.Errorf("invalid interval %d-%d", from, to)
	}

	result := make([]int, 0, l0)
	for i := from; i <= to; i++ {
		result = append(result, i)
	}

	return result, nil
}

func newTopology(cpuTopology *sysfs.CPUTopology) (*Topology, error) {
	coreID, err := strconv.Atoi(cpuTopology.CoreID)
	if err != nil {
		log.Errorf("newTopology, strconv.Atoi(cpuTopology.CoreID) err = %v", err)
		return nil, err
	}

	coreSiblingsList, err := parseIntList0(cpuTopology.CoreSiblingsList)
	if err != nil {
		log.Errorf("newTopology, parseIntList(cpuTopology.CoreSiblingsList) err = %v", err)
		return nil, err
	}

	physicalPackageID, err := strconv.Atoi(cpuTopology.PhysicalPackageID)
	if err != nil {
		log.Errorf("newTopology, strconv.Atoi(cpuTopology.PhysicalPackageID) err = %v", err)
		return nil, err
	}

	threadSiblingsList, err := parseIntList(cpuTopology.ThreadSiblingsList)
	if err != nil {
		log.Errorf("newTopology, parseIntList(cpuTopology.ThreadSiblingsList) err = %v", err)
		return nil, err
	}

	topology := &Topology{
		CoreID:             coreID,
		CoreSiblingsList:   coreSiblingsList,
		PhysicalPackageID:  physicalPackageID,
		ThreadSiblingsList: threadSiblingsList,
	}
	return topology, nil
}
