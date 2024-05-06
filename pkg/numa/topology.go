package numa

import (
	"fmt"
	"github.com/prometheus/procfs/sysfs"
	log "github.com/sirupsen/logrus"
	"sort"
	"strconv"
	"strings"
)

type CpuTopologyInfo struct {
	Cpu         sysfs.CPU
	CpuNumber   string
	Number      int                // like CpuNumber, but int
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
			info.CpuNumber, tpg.CoreID, tpg.CoreSiblingsList, tpg.PhysicalPackageID,
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
		log.Errorf("GetCpuTopologyInfo, sysfs0.CPUs err = %v", err)
		return nil, err
	}

	l0 := len(cpus0)
	infos := make([]*CpuTopologyInfo, 0, l0)
	for i := 0; i < l0; i++ {
		cpu0 := cpus0[i]
		cpuNumber := cpu0.Number()
		cpuTopology, err := cpu0.Topology()
		if err != nil {
			log.Errorf("GetCpuTopologyInfo, cpu0.Topology err = %v", err)
			return nil, err
		}

		number, err := strconv.Atoi(cpuNumber)
		if err != nil {
			log.Errorf("GetCpuTopologyInfo, strconv.Atoi(cpuNumber) err = %v, cpuNumber = %s", err, cpuNumber)
			return nil, err
		}

		topology, err := newTopology(cpuTopology)
		if err != nil {
			log.Errorf("GetCpuTopologyInfo, newTopology err = %v", err)
			return nil, err
		}

		ti := &CpuTopologyInfo{
			CpuNumber:   cpuNumber,
			Number:      number,
			Cpu:         cpu0,
			CpuTopology: cpuTopology,
			Topology:    topology,
		}
		infos = append(infos, ti)
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Number < infos[j].Number
	})

	return infos, nil
}

// parseIntList - parse list like "1,7,8,16"
// and like "0-15"
// and like "0-23,48-71"
func parseIntList(val0 string) ([]int, error) {
	comma := strings.Contains(val0, ",")
	dash := strings.Contains(val0, "-")
	if comma && dash {
		return parseIntList2(val0)
	}
	if comma {
		return parseIntList0(val0)
	}
	return parseIntList1(val0)
}

// parseIntList0 - parse list like "1,7,8,16"
func parseIntList0(val0 string) ([]int, error) {
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

// parseIntList0 - parse list like "0-15"
func parseIntList1(val0 string) ([]int, error) {
	split0 := strings.Split(val0, "-")
	l0 := len(split0)
	if l0 != 2 {
		//log.Errorf("parseIntList1, invalid split count = %d", l0)
		//return nil, fmt.Errorf("invalid split count = %d", l0)
		return nil, nil
	}

	str0 := strings.TrimSpace(split0[0])
	from, err := strconv.Atoi(str0)
	if err != nil {
		log.Errorf("parseIntList1, strconv.Atoi(str0) err = %v, str0 = %v", err, str0)
		return nil, err
	}

	str1 := strings.TrimSpace(split0[1])
	to, err := strconv.Atoi(str1)
	if err != nil {
		log.Errorf("parseIntList1, strconv.Atoi(str1) err = %v, str1 = %v", err, str1)
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

// parseIntList2 - parse list like "0-23,48-71"
func parseIntList2(val0 string) ([]int, error) {
	split0 := strings.Split(val0, ",")
	l0 := len(split0)
	var result []int
	for i := 0; i < l0; i++ {
		str0 := strings.TrimSpace(split0[i]) // 0-23
		sl0, err := parseIntList1(str0)
		if err != nil {
			return nil, err
		}
		result = append(result, sl0...)
	}
	return result, nil
}

func newTopology(cpuTopology *sysfs.CPUTopology) (*Topology, error) {
	coreID, err := strconv.Atoi(cpuTopology.CoreID)
	if err != nil {
		log.Errorf("newTopology, strconv.Atoi(cpuTopology.CoreID) err = %v, cpuTopology.CoreID = %s", err, cpuTopology.CoreID)
		return nil, err
	}

	coreSiblingsList, err := parseIntList(cpuTopology.CoreSiblingsList)
	if err != nil {
		log.Errorf("newTopology, parseIntList(cpuTopology.CoreSiblingsList) err = %v, cpuTopology.CoreSiblingsList = %s", err, cpuTopology.CoreSiblingsList)
		return nil, err
	}

	physicalPackageID, err := strconv.Atoi(cpuTopology.PhysicalPackageID)
	if err != nil {
		log.Errorf("newTopology, strconv.Atoi(cpuTopology.PhysicalPackageID) err = %v, cpuTopology.PhysicalPackageID = %s", err, cpuTopology.PhysicalPackageID)
		return nil, err
	}

	threadSiblingsList, err := parseIntList(cpuTopology.ThreadSiblingsList)
	if err != nil {
		log.Errorf("newTopology, parseIntList(cpuTopology.ThreadSiblingsList) err = %v, cpuTopology.ThreadSiblingsList = %s", err, cpuTopology.ThreadSiblingsList)
		return nil, err
	}
	sort.Slice(threadSiblingsList, func(i, j int) bool {
		return threadSiblingsList[i] < threadSiblingsList[j]
	})

	topology := &Topology{
		CoreID:             coreID,
		CoreSiblingsList:   coreSiblingsList,
		PhysicalPackageID:  physicalPackageID,
		ThreadSiblingsList: threadSiblingsList,
	}
	return topology, nil
}
