package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type StaticCollector struct {
	IdleLoad0 *prometheus.Desc
	IdleLoad1 *prometheus.Desc
	LoadFree0 *prometheus.Desc
	LoadFree1 *prometheus.Desc
	LoadUsed0 *prometheus.Desc
	LoadUsed1 *prometheus.Desc

	State *State
}

func NewStaticCollector() *StaticCollector {
	ent := &StaticCollector{}

	ent.IdleLoad0 = prometheus.NewDesc("idle_load0",
		"400, 600, 800 %",
		nil, nil,
	)
	ent.IdleLoad1 = prometheus.NewDesc("idle_load1",
		"0-100 %",
		nil, nil,
	)
	ent.LoadFree0 = prometheus.NewDesc("load_free0",
		"400, 600, 800 %",
		nil, nil,
	)
	ent.LoadFree1 = prometheus.NewDesc("load_free1",
		"0-100 %",
		nil, nil,
	)
	ent.LoadUsed0 = prometheus.NewDesc("load_used0",
		"400, 600, 800 %",
		nil, nil,
	)
	ent.LoadUsed1 = prometheus.NewDesc("load_used1",
		"0-100 %",
		nil, nil,
	)

	ent.State = NewState()

	return ent
}

func (x *StaticCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- x.IdleLoad0
	ch <- x.IdleLoad1
}

func (x *StaticCollector) Collect(ch chan<- prometheus.Metric) {
	m0 := prometheus.MustNewConstMetric(x.IdleLoad0, prometheus.GaugeValue, x.State.IdleLoad0)
	m1 := prometheus.MustNewConstMetric(x.IdleLoad1, prometheus.GaugeValue, x.State.IdleLoad1)
	m2 := prometheus.MustNewConstMetric(x.LoadFree0, prometheus.GaugeValue, x.State.LoadFree0)
	m3 := prometheus.MustNewConstMetric(x.LoadFree1, prometheus.GaugeValue, x.State.LoadFree1)
	m4 := prometheus.MustNewConstMetric(x.LoadUsed0, prometheus.GaugeValue, x.State.LoadUsed0)
	m5 := prometheus.MustNewConstMetric(x.LoadUsed1, prometheus.GaugeValue, x.State.LoadUsed1)

	now := time.Now()
	mt0 := prometheus.NewMetricWithTimestamp(now, m0)
	mt1 := prometheus.NewMetricWithTimestamp(now, m1)
	mt2 := prometheus.NewMetricWithTimestamp(now, m2)
	mt3 := prometheus.NewMetricWithTimestamp(now, m3)
	mt4 := prometheus.NewMetricWithTimestamp(now, m4)
	mt5 := prometheus.NewMetricWithTimestamp(now, m5)

	ch <- mt0
	ch <- mt1
	ch <- mt2
	ch <- mt3
	ch <- mt4
	ch <- mt5
}
