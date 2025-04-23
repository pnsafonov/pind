package monitoring

import (
	"fmt"
	"github.com/pnsafonov/pind/pkg/config"
	"github.com/pnsafonov/pind/pkg/monitoring/collector"
	"github.com/pnsafonov/pind/pkg/monitoring/mon_state"
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
	staticCollector *collector.Static
	procCollectors  map[string]*collector.Proc
}

func NewMonitoring(config *config.Monitoring, numaNodesCount int) *Monitoring {
	static0 := collector.NewStatic(numaNodesCount)

	ent := &Monitoring{
		config:          config,
		staticCollector: static0,
		procCollectors:  make(map[string]*collector.Proc),
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
	mux.Handle("/metrics", handler)
	x.server = &http.Server{
		Handler: mux,
	}

	go x.serve(listener)
	return nil
}

func (x *Monitoring) SetState(state *mon_state.State) {
	x.staticCollector.State = state

	// для процесса регистрируем метрику
	for _, proc := range state.Procs {
		_ = x.setProc(proc)
	}

	// процесс завершился, убираем метрику
	for vmName, procCollector := range x.procCollectors {
		_, ok := state.Procs[vmName]
		if !ok {
			_ = x.registerer.Unregister(procCollector)
			delete(x.procCollectors, vmName)
		}
	}
}

func (x *Monitoring) setProc(proc *mon_state.Proc) error {
	if proc.VmName == "" {
		return fmt.Errorf("proc.VmName is empty")
	}

	var err error
	procCollector, ok := x.procCollectors[proc.VmName]
	if !ok {
		procCollector, err = collector.NewProc(proc)
		if err != nil {
			return err
		}

		err = x.registerer.Register(procCollector)
		if err != nil {
			log.Errorf("setProc, prometheus.Register err: %v", err)
			return err
		}

		x.procCollectors[proc.VmName] = procCollector
	}

	procCollector.SetProc(proc)
	return nil
}
