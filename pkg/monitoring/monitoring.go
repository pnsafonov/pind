package monitoring

import (
	"github.com/pnsafonov/pind/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
)

type Monitoring struct {
	config          *config.Monitoring
	server          *http.Server
	gatherer        prometheus.Gatherer
	registerer      prometheus.Registerer
	staticCollector *StaticCollector
}

func NewMonitoring(config *config.Monitoring) *Monitoring {
	collector := NewStaticCollector()

	ent := &Monitoring{
		config:          config,
		staticCollector: collector,
	}
	return ent
}

func (x *Monitoring) serve(l net.Listener) {
	_ = x.server.Serve(l)
}

func (x *Monitoring) GoServe() error {
	addr := x.config.Listen

	// включение Go-шных метрик
	if x.config.GoMetricsEnabled {
		x.gatherer = prometheus.DefaultGatherer
		x.registerer = prometheus.DefaultRegisterer
	} else {
		reg := prometheus.NewRegistry()
		x.gatherer = reg
		x.registerer = reg
	}

	err := x.registerer.Register(x.staticCollector)
	if err != nil {
		log.Errorf("GoServe, prometheus.Register err: %v", err)
		return err
	}

	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Errorf("ListenAndServe, net.Listen err = %v", err)
		return err
	}

	handlerOptions := promhttp.HandlerOpts{
		//EnableOpenMetrics: false,
	}
	handler := promhttp.HandlerFor(x.gatherer, handlerOptions)

	mux := http.NewServeMux()
	mux.Handle("/console/metrics", handler)
	x.server = &http.Server{
		Handler: mux,
	}

	go x.serve(listener)
	return nil
}

func (x *Monitoring) SetState(state *State) {
	x.staticCollector.State = state
}
