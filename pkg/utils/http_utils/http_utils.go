package http_utils

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type JsonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func GetRequestParam0(r *http.Request, keys []string) string {
	query := r.URL.Query()
	for _, key := range keys {
		val := query.Get(key)
		if val != "" {
			return val
		}
	}
	return ""
}

func WriteResponse0(w http.ResponseWriter, code int, message string) {
	jsonResponse := &JsonResponse{
		Code:    code,
		Message: message,
	}
	bytes0, err := json.MarshalIndent(jsonResponse, "", "    ")
	if err != nil {
		log.Errorf("WriteResponse0, json.MarshalIndent err = %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(bytes0)
	w.WriteHeader(code)
}

func WriteResponse1(w http.ResponseWriter, data interface{}) error {
	bytes0, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Errorf("WriteResponse10 json.Marshal err = %v", err)
		msg := fmt.Sprintf("Internal Server Error %v", err)
		WriteResponse0(w, http.StatusInternalServerError, msg)
		return err
	}

	_, err = w.Write(bytes0)
	return err
}

func WriteResponse2(w http.ResponseWriter, data interface{}) {
	_ = WriteResponse1(w, data)
}

func ApplicationJsonHeader(w http.ResponseWriter) {
	header := w.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
}
