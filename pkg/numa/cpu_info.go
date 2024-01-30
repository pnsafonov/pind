package numa

import (
	"github.com/prometheus/procfs"
	log "github.com/sirupsen/logrus"
)

type Info struct {
	Nodes []*NodeInfo
	Infos []*NodeCpuInfo
}

type NodeCpuInfo struct {
	Index int
	Info  *NodeInfo
	Cores map[int]*CpuInfo
}

type CpuInfo struct {
	CPU     int            // cpu number in system
	Stat    procfs.CPUStat // content of /proc/stat for cpu
	CpuLoad float64        // % of cpu time used
}

func (x *Info) GetCpuInfo(cpu int) (*CpuInfo, bool) {
	l0 := len(x.Infos)
	for i := 0; i < l0; i++ {
		info := x.Infos[i]
		cpuInfo, ok := info.Cores[cpu]
		if ok {
			return cpuInfo, true
		}
	}
	return nil, false
}

func newInfo() (*Info, error) {
	if initError != nil {
		return nil, initError
	}
	info := &Info{
		Nodes: nodes,
		Infos: make([]*NodeCpuInfo, 0, len(nodes)),
	}
	return info, nil
}

// GetSumm0 - returns all except idle
func (x *CpuInfo) GetSumm0() float64 {
	summ0 := float64(0)

	summ0 += x.Stat.User
	summ0 += x.Stat.Nice
	summ0 += x.Stat.System
	//summ0 += x.Stat.Idle
	summ0 += x.Stat.Iowait
	summ0 += x.Stat.IRQ
	summ0 += x.Stat.SoftIRQ
	summ0 += x.Stat.Steal
	summ0 += x.Stat.Guest
	summ0 += x.Stat.GuestNice

	return summ0
}

func (x *CpuInfo) GetSumm() float64 {
	summ := x.GetSumm0()
	summ += x.Stat.Idle
	return summ
}

func GetCpuInfos() (*Info, error) {
	procStat, err := procfs0.Stat()
	if err != nil {
		log.Errorf("GetCpuInfos, procfs0.Stat err = %v", err)
		return nil, err
	}

	info, err := newInfo()
	if err != nil {
		log.Errorf("GetCpuInfos, newInfo err = %v", err)
		return nil, err
	}

	for i, node := range info.Nodes {
		nodeCpuInfo := &NodeCpuInfo{
			Index: i,
			Info:  node,
			Cores: make(map[int]*CpuInfo),
		}

		for _, cpu := range node.Cpus {
			stat, ok := procStat.CPU[int64(cpu)]
			if !ok {
				continue
			}

			cpuInfo := &CpuInfo{
				CPU:  cpu,
				Stat: stat,
			}

			nodeCpuInfo.Cores[cpu] = cpuInfo
		}

		info.Infos = append(info.Infos, nodeCpuInfo)
	}

	return info, nil
}
