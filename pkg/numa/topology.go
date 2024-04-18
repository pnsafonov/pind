package numa

import (
	"fmt"
	"github.com/prometheus/procfs/sysfs"
	log "github.com/sirupsen/logrus"
)

type CpuTopologyInfo struct {
	Cpu      sysfs.CPU
	Number   string
	Topology *sysfs.CPUTopology
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
		tpg := info.Topology
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
		topology, err := cpu0.Topology()
		if err != nil {
			log.Errorf("PrintTopology, cpu0.Topology err = %v", err)
			return nil, err
		}

		ti := &CpuTopologyInfo{
			Number:   number,
			Cpu:      cpu0,
			Topology: topology,
		}
		infos = append(infos, ti)
	}
	return infos, nil
}
