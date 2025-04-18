package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type StaticCollector struct {
	IdleLoad0 *prometheus.Desc
	IdleLoad1 *prometheus.Desc

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

	now := time.Now()
	mt0 := prometheus.NewMetricWithTimestamp(now, m0)
	mt1 := prometheus.NewMetricWithTimestamp(now, m1)

	ch <- mt0
	ch <- mt1
}
