package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type resp struct {
	Msg     string      `json:"msg"`
	Success bool        `json:"success"`
	Payload interface{} `json:"payload"`
}

func jsonResp(w http.ResponseWriter, header int, r resp) {
	respObj, err := json.Marshal(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf("error: could not marshal provided payload: %s", r.Payload)
		w.Write([]byte(msg))
	} else {
		w.WriteHeader(header)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respObj)
	}
}
