package collector

import (
	"fmt"
	"github.com/pnsafonov/pind/pkg/monitoring/mon_state"
	"github.com/prometheus/client_golang/prometheus"
)

type Proc struct {
	VmName string
	Proc   *mon_state.Proc

	CPU   *prometheus.Desc
	Load  *prometheus.Desc
	Numa0 *prometheus.Desc
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
	m0 := prometheus.MustNewConstMetric(x.CPU, prometheus.GaugeValue, x.Proc.CPU)
	m1 := prometheus.MustNewConstMetric(x.Load, prometheus.GaugeValue, x.Proc.GetLoad0())
	m2 := prometheus.MustNewConstMetric(x.Numa0, prometheus.GaugeValue, float64(x.Proc.Numa0))

	t0 := x.Proc.Time
	mt0 := prometheus.NewMetricWithTimestamp(t0, m0)
	mt1 := prometheus.NewMetricWithTimestamp(t0, m1)
	mt2 := prometheus.NewMetricWithTimestamp(t0, m2)

	ch <- mt0
	ch <- mt1
	ch <- mt2
}
