package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

type StaticCollector struct {
	PoolCollector *PoolCollector

	State *State
}

type PoolCollector struct {
	IdleLoad0 *prometheus.Desc
	IdleLoad1 *prometheus.Desc
	LoadFree0 *prometheus.Desc
	LoadFree1 *prometheus.Desc
	LoadUsed0 *prometheus.Desc
	LoadUsed1 *prometheus.Desc

	Nodes []*PoolNodeCollector
}

type PoolNodeCollector struct {
	Index     int
	LoadFree0 *prometheus.Desc
	LoadFree1 *prometheus.Desc
	LoadUsed0 *prometheus.Desc
	LoadUsed1 *prometheus.Desc
}

func NewPoolCollector(nodesCount int) *PoolCollector {
	ent := &PoolCollector{}

	ent.IdleLoad0 = prometheus.NewDesc("pind_pool_idle_load0",
		"idle cores load, like 400, 600, 800 %",
		nil, nil,
	)
	ent.IdleLoad1 = prometheus.NewDesc("pind_pool_idle_load1",
		"idle cores load, like 0-100 %",
		nil, nil,
	)
	ent.LoadFree0 = prometheus.NewDesc("pind_pool_load_free0",
		"free cores load, like 400, 600, 800 %",
		nil, nil,
	)
	ent.LoadFree1 = prometheus.NewDesc("pind_pool_load_free1",
		"free cores load, like 0-100 %",
		nil, nil,
	)
	ent.LoadUsed0 = prometheus.NewDesc("pind_pool_load_used0",
		"used cores load, like 400, 600, 800 %",
		nil, nil,
	)
	ent.LoadUsed1 = prometheus.NewDesc("pind_pool_load_used1",
		"used cores load, like 0-100 %",
		nil, nil,
	)

	for i := 0; i < nodesCount; i++ {
		node := &PoolNodeCollector{
			Index: i,
		}

		labels0 := prometheus.Labels{
			"node": strconv.Itoa(i),
		}

		node.LoadFree0 = prometheus.NewDesc("pind_node_load_free0",
			"node free cores load, like 400, 600, 800 %",
			nil, labels0,
		)
		node.LoadFree1 = prometheus.NewDesc("pind_node_load_free1",
			"node free cores load, like 0-100 %",
			nil, labels0,
		)
		node.LoadUsed0 = prometheus.NewDesc("pind_node_load_used0",
			"node used cores load, like 400, 600, 800 %",
			nil, labels0,
		)
		node.LoadUsed1 = prometheus.NewDesc("pind_node_load_used1",
			"node used cores load, like 0-100 %",
			nil, labels0,
		)

		ent.Nodes = append(ent.Nodes, node)
	}

	return ent
}

func NewStaticCollector(nodesCount int) *StaticCollector {
	pool := NewPoolCollector(nodesCount)

	ent := &StaticCollector{
		PoolCollector: pool,
	}

	return ent
}

func (x *StaticCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- x.PoolCollector.IdleLoad0
	ch <- x.PoolCollector.IdleLoad1
	ch <- x.PoolCollector.LoadFree0
	ch <- x.PoolCollector.LoadFree1
	ch <- x.PoolCollector.LoadUsed0
	ch <- x.PoolCollector.LoadUsed1

	for _, node := range x.PoolCollector.Nodes {
		ch <- node.LoadFree0
		ch <- node.LoadFree1
		ch <- node.LoadUsed0
		ch <- node.LoadUsed1
	}

}

func (x *StaticCollector) Collect(ch chan<- prometheus.Metric) {
	m0 := prometheus.MustNewConstMetric(x.PoolCollector.IdleLoad0, prometheus.GaugeValue, x.State.Pool.IdleLoad0)
	m1 := prometheus.MustNewConstMetric(x.PoolCollector.IdleLoad1, prometheus.GaugeValue, x.State.Pool.IdleLoad1)
	m2 := prometheus.MustNewConstMetric(x.PoolCollector.LoadFree0, prometheus.GaugeValue, x.State.Pool.LoadFree0)
	m3 := prometheus.MustNewConstMetric(x.PoolCollector.LoadFree1, prometheus.GaugeValue, x.State.Pool.LoadFree1)
	m4 := prometheus.MustNewConstMetric(x.PoolCollector.LoadUsed0, prometheus.GaugeValue, x.State.Pool.LoadUsed0)
	m5 := prometheus.MustNewConstMetric(x.PoolCollector.LoadUsed1, prometheus.GaugeValue, x.State.Pool.LoadUsed1)

	t0 := x.State.Time
	mt0 := prometheus.NewMetricWithTimestamp(t0, m0)
	mt1 := prometheus.NewMetricWithTimestamp(t0, m1)
	mt2 := prometheus.NewMetricWithTimestamp(t0, m2)
	mt3 := prometheus.NewMetricWithTimestamp(t0, m3)
	mt4 := prometheus.NewMetricWithTimestamp(t0, m4)
	mt5 := prometheus.NewMetricWithTimestamp(t0, m5)

	ch <- mt0
	ch <- mt1
	ch <- mt2
	ch <- mt3
	ch <- mt4
	ch <- mt5

	l0 := len(x.PoolCollector.Nodes)
	for i := 0; i < l0; i++ {
		node := x.State.Pool.Nodes[i]

		nodeCollector := x.PoolCollector.Nodes[i]

		mn0 := prometheus.MustNewConstMetric(nodeCollector.LoadFree0, prometheus.GaugeValue, node.LoadFree0)
		mn1 := prometheus.MustNewConstMetric(nodeCollector.LoadFree1, prometheus.GaugeValue, node.LoadFree1)
		mn2 := prometheus.MustNewConstMetric(nodeCollector.LoadUsed0, prometheus.GaugeValue, node.LoadUsed0)
		mn3 := prometheus.MustNewConstMetric(nodeCollector.LoadUsed1, prometheus.GaugeValue, node.LoadUsed1)

		mnt0 := prometheus.NewMetricWithTimestamp(t0, mn0)
		mnt1 := prometheus.NewMetricWithTimestamp(t0, mn1)
		mnt2 := prometheus.NewMetricWithTimestamp(t0, mn2)
		mnt3 := prometheus.NewMetricWithTimestamp(t0, mn3)

		ch <- mnt0
		ch <- mnt1
		ch <- mnt2
		ch <- mnt3
	}
}
