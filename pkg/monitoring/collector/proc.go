package collector

import (
	"fmt"
	"github.com/pnsafonov/pind/pkg/monitoring/mon_state"
	"github.com/prometheus/client_golang/prometheus"
)

type Proc struct {
	VmName string
	Proc   *mon_state.Proc

	CPU               *prometheus.Desc
	Load              *prometheus.Desc
	Numa0             *prometheus.Desc
	RequiredCoresPhys *prometheus.Desc
	RequiredCores     *prometheus.Desc
	AssignedCores     *prometheus.Desc
	AssignedCores0    *prometheus.Desc
}

func NewProc(proc *mon_state.Proc) (*Proc, error) {
	vmName := proc.VmName
	if vmName == "" {
		return nil, fmt.Errorf("vm name is required")
	}

	ent := &Proc{
		VmName: vmName,
		Proc:   proc,
	}

	labels0 := prometheus.Labels{
		"vm_name": vmName,
	}

	ent.CPU = prometheus.NewDesc("pind_vm_cpu",
		"vm cpu",
		nil, labels0,
	)
	ent.Load = prometheus.NewDesc("pind_vm_load",
		"is vm on load",
		nil, labels0,
	)
	ent.Numa0 = prometheus.NewDesc("pind_vm_numa",
		"assigned numa for vm, -1 for not assigned",
		nil, labels0,
	)

	ent.RequiredCoresPhys = prometheus.NewDesc("pind_vm_required_cores_phys",
		"vm required physical cpu cores",
		nil, labels0,
	)
	ent.RequiredCores = prometheus.NewDesc("pind_vm_required_cores",
		"vm required cpu cores",
		nil, labels0,
	)
	ent.AssignedCores = prometheus.NewDesc("pind_vm_assigned_cores",
		"vm assigned cpu cores",
		nil, labels0,
	)
	ent.AssignedCores0 = prometheus.NewDesc("pind_vm_assigned_cores0",
		"vm assigned cpu cores in percents, max 100%",
		nil, labels0,
	)

	return ent, nil
}

func (x *Proc) SetProc(proc *mon_state.Proc) {
	x.Proc = proc
}

func (x *Proc) Describe(ch chan<- *prometheus.Desc) {
	ch <- x.CPU
	ch <- x.Load
	ch <- x.Numa0
}

func (x *Proc) Collect(ch chan<- prometheus.Metric) {
	assignedCores0 := float64(0)
	if x.Proc.RequiredCores != 0 {
		assignedCores0 = (float64(x.Proc.AssignedCores) / float64(x.Proc.RequiredCores)) * 100
	}

	m0 := prometheus.MustNewConstMetric(x.CPU, prometheus.GaugeValue, x.Proc.CPU)
	m1 := prometheus.MustNewConstMetric(x.Load, prometheus.GaugeValue, x.Proc.GetLoad0())
	m2 := prometheus.MustNewConstMetric(x.Numa0, prometheus.GaugeValue, float64(x.Proc.Numa0))
	m3 := prometheus.MustNewConstMetric(x.RequiredCoresPhys, prometheus.GaugeValue, float64(x.Proc.RequiredCoresPhys))
	m4 := prometheus.MustNewConstMetric(x.RequiredCores, prometheus.GaugeValue, float64(x.Proc.RequiredCores))
	m5 := prometheus.MustNewConstMetric(x.AssignedCores, prometheus.GaugeValue, float64(x.Proc.AssignedCores))
	m6 := prometheus.MustNewConstMetric(x.AssignedCores0, prometheus.GaugeValue, assignedCores0)

	t0 := x.Proc.Time
	mt0 := prometheus.NewMetricWithTimestamp(t0, m0)
	mt1 := prometheus.NewMetricWithTimestamp(t0, m1)
	mt2 := prometheus.NewMetricWithTimestamp(t0, m2)
	mt3 := prometheus.NewMetricWithTimestamp(t0, m3)
	mt4 := prometheus.NewMetricWithTimestamp(t0, m4)
	mt5 := prometheus.NewMetricWithTimestamp(t0, m5)
	mt6 := prometheus.NewMetricWithTimestamp(t0, m6)

	ch <- mt0
	ch <- mt1
	ch <- mt2
	ch <- mt3
	ch <- mt4
	ch <- mt5
	ch <- mt6
}
