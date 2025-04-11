package http_api

import (
	"encoding/json"
	"fmt"
	"github.com/pnsafonov/pind/pkg/ali"
	"github.com/pnsafonov/pind/pkg/config"
	"github.com/pnsafonov/pind/pkg/utils/http_utils"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strconv"
)

type HttpApi struct {
	Config *config.HttpApi
	Server *http.Server
	Logic  ali.ApiLogic

	state      *State
	stateBytes []byte
}

func NewHttpApi(logic ali.ApiLogic, config0 *config.HttpApi) *HttpApi {
	httpApi := &HttpApi{
		Config: config0,
		Logic:  logic,
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
	mux.HandleFunc("/", x.getApiState)
	mux.HandleFunc("/api/state", x.getApiState)
	mux.HandleFunc("/api/pin", x.pin)
	mux.HandleFunc("/api/pin_add", x.pin)
	mux.HandleFunc("/api/pin/add", x.pin)
	mux.HandleFunc("/api/pin_list", x.pinList)
	mux.HandleFunc("/api/pin/list", x.pinList)
	mux.HandleFunc("/api/pin_clean", x.pinClean)
	mux.HandleFunc("/api/pin/clean", x.pinClean)
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

func writeStateBytes(w http.ResponseWriter, state *State) error {
	stateBytes, err := json.MarshalIndent(state, "", "    ")
	if err != nil {
		log.Errorf("HttpApi writeStateBytes, json.Marshal err = %v", err)
		return err
	}

	_, err = w.Write(stateBytes)
	return err
}

// getApiState - /api/state
func (x *HttpApi) getApiState(w http.ResponseWriter, r *http.Request) {
	http_utils.ApplicationJsonHeader(w)

	// /api/state?vm_name=my_env_db
	vmName := r.URL.Query().Get("vm_name")
	if vmName != "" {
		state1 := x.state.CloneWithFilter1(vmName)
		_ = writeStateBytes(w, state1)
		return
	}

	// /api/state?vm_prefix=my_env_
	vmPrefix := r.URL.Query().Get("vm_prefix")
	if vmPrefix != "" {
		state1 := x.state.CloneWithFilter2(vmPrefix)
		_ = writeStateBytes(w, state1)
		return
	}

	// /api/state
	_, _ = w.Write(x.stateBytes)
}

var (
	keysVmName    = []string{"vm_name", "name"}
	keysNumaIndex = []string{"numa", "numa_id", "numa_index"}
)

// getApiState - /api/pin, /api/pin_add, /api/pin/add
func (x *HttpApi) pin(w http.ResponseWriter, r *http.Request) {
	http_utils.ApplicationJsonHeader(w)

	vmName := http_utils.GetRequestParam0(r, keysVmName)
	if vmName == "" {
		msg := fmt.Sprintf("need to specify argument name with %v", keysVmName)
		http_utils.WriteResponse0(w, http.StatusBadRequest, msg)
		return
	}

	numa0 := http_utils.GetRequestParam0(r, keysNumaIndex)
	if numa0 == "" {
		msg := fmt.Sprintf("need to specify argument numa with %v", keysNumaIndex)
		http_utils.WriteResponse0(w, http.StatusBadRequest, msg)
		return
	}

	numa1, err := strconv.Atoi(numa0)
	if err != nil {
		log.Errorf("HttpApi pin, strconv.Atoi err = %v", err)
		msg := fmt.Sprintf("can't parse argument")
		http_utils.WriteResponse0(w, http.StatusBadRequest, msg)
		return
	}
	if numa1 < 0 {
		log.Errorf("HttpApi pin, negative numa = %d", numa0)
		msg := fmt.Sprintf("numa is negative, numa = %d", numa0)
		http_utils.WriteResponse0(w, http.StatusBadRequest, msg)
		return
	}

	err = x.Logic.AddPinMapping(vmName, numa1)
	if err != nil {
		msg := fmt.Sprintf("AddPinMapping, err = %v", err)
		http_utils.WriteResponse0(w, http.StatusBadRequest, msg)
		return
	}

	msg := fmt.Sprintf("pin, added name = %s, numa = %d", vmName, numa1)
	http_utils.WriteResponse0(w, http.StatusOK, msg)
}

// pinList - /api/pin_list, /api/pin/list
func (x *HttpApi) pinList(w http.ResponseWriter, r *http.Request) {
	http_utils.ApplicationJsonHeader(w)

	pinMapping := x.Logic.GetPinMapping()
	http_utils.WriteResponse2(w, pinMapping)
}

// pinRemove - /api/pin_remove, /api/pin/remove
func (x *HttpApi) pinRemove(w http.ResponseWriter, r *http.Request) {
	http_utils.ApplicationJsonHeader(w)

	vmName := http_utils.GetRequestParam0(r, keysVmName)
	if vmName == "" {
		msg := fmt.Sprintf("need to specify argument name with %v", keysVmName)
		http_utils.WriteResponse0(w, http.StatusBadRequest, msg)
		return
	}

	x.Logic.Remove(vmName)

	msg := fmt.Sprintf("pinRemove, removed name = %s, numa = %d", vmName)
	http_utils.WriteResponse0(w, http.StatusOK, msg)
}

// pinClean - /api/pin_clean, /api/pin/clean
func (x *HttpApi) pinClean(w http.ResponseWriter, r *http.Request) {
	http_utils.ApplicationJsonHeader(w)

	x.Logic.Clean()

	msg := fmt.Sprintf("pinClean is ok")
	http_utils.WriteResponse0(w, http.StatusOK, msg)
}
