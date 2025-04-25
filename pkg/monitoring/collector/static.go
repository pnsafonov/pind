package collector

import (
	"github.com/pnsafonov/pind/pkg/monitoring/mon_state"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

// Static - статический Collector, регистрируется при старте
// далее для него не делается unregister
type Static struct {
	PoolCollector *Pool

	State *mon_state.State
}

type Pool struct {
	IdleLoad0 *prometheus.Desc
	IdleLoad1 *prometheus.Desc
	LoadFree0 *prometheus.Desc
	LoadFree1 *prometheus.Desc
	LoadUsed0 *prometheus.Desc
	LoadUsed1 *prometheus.Desc

	Nodes []*PoolNode
}

type PoolNode struct {
	Index         int
	LoadFree0     *prometheus.Desc
	LoadFree1     *prometheus.Desc
	LoadUsed0     *prometheus.Desc
	LoadUsed1     *prometheus.Desc
	LoadFreeCount *prometheus.Desc
	LoadUsedCount *prometheus.Desc
}

func NewPool(nodesCount int) *Pool {
	ent := &Pool{}

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
		node := &PoolNode{
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
		node.LoadFreeCount = prometheus.NewDesc("pind_node_load_free_count",
			"free cpu cores count of numa node",
			nil, labels0,
		)
		node.LoadUsedCount = prometheus.NewDesc("pind_node_load_used_count",
			"used cpu cores count of numa node",
			nil, labels0,
		)

		ent.Nodes = append(ent.Nodes, node)
	}

	return ent
}

func NewStatic(nodesCount int) *Static {
	pool := NewPool(nodesCount)

	ent := &Static{
		PoolCollector: pool,
	}

	return ent
}

func (x *Static) Describe(ch chan<- *prometheus.Desc) {
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

func (x *Static) Collect(ch chan<- prometheus.Metric) {
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
		mn4 := prometheus.MustNewConstMetric(nodeCollector.LoadFreeCount, prometheus.GaugeValue, node.LoadFreeCount)
		mn5 := prometheus.MustNewConstMetric(nodeCollector.LoadUsedCount, prometheus.GaugeValue, node.LoadUsedCount)

		mnt0 := prometheus.NewMetricWithTimestamp(t0, mn0)
		mnt1 := prometheus.NewMetricWithTimestamp(t0, mn1)
		mnt2 := prometheus.NewMetricWithTimestamp(t0, mn2)
		mnt3 := prometheus.NewMetricWithTimestamp(t0, mn3)
		mnt4 := prometheus.NewMetricWithTimestamp(t0, mn4)
		mnt5 := prometheus.NewMetricWithTimestamp(t0, mn5)

		ch <- mnt0
		ch <- mnt1
		ch <- mnt2
		ch <- mnt3
		ch <- mnt4
		ch <- mnt5
	}
}
