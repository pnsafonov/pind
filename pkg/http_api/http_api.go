package http_api

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"pind/pkg/config"
)

type HttpApi struct {
	Config *config.HttpApi
	Server *http.Server

	state      *State
	stateBytes []byte
}

func NewHttpApi(config0 *config.HttpApi) *HttpApi {
	httpApi := &HttpApi{
		Config: config0,
	}
	httpApi.setState0()
	return httpApi
}

func (x *HttpApi) GoServe() error {
	addr := x.Config.Listen

	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Errorf("ListenAndServe, net.Listen err = %v", err)
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/state", x.getApiState)
	x.Server = &http.Server{
		Handler: mux,
	}

	go x.serve(listener)
	return nil
}

func (x *HttpApi) SetState(state *State) error {
	stateBytes, err := json.MarshalIndent(state, "", "    ")
	if err != nil {
		log.Errorf("HttpApi SetState, json.Marshal err = %v", err)
		return err
	}

	x.state = state
	x.stateBytes = stateBytes
	return nil
}

func (x *HttpApi) setState0() {
	state := &State{}
	_ = x.SetState(state)
}

func (x *HttpApi) serve(l net.Listener) {
	_ = x.Server.Serve(l)
}

// getApiState - /api/state
func (x *HttpApi) getApiState(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")

	_, _ = w.Write(x.stateBytes)
}
