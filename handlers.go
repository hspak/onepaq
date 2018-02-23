package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

// TODO: setup middlewares so we can remove some boilerplate.

type passPayload struct {
	Password string `json:"password"`
}

func (s *server) StatusHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if s.timerStartTime.IsZero() {
		jsonResp(w, http.StatusBadRequest, resp{Msg: "vault is locked", Success: false, Payload: nil})
		return
	}
	since := s.timeout - time.Since(s.timerStartTime)
	msg := fmt.Sprintf("vault is unlocked, %s until vault is autolocked", since)
	jsonResp(w, http.StatusOK, resp{Msg: msg, Success: true, Payload: nil})
}

func (s *server) ItemsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if s.timerStartTime.IsZero() {
		jsonResp(w, http.StatusBadRequest, resp{Msg: "vault is locked", Success: false, Payload: nil})
		return
	}
	titles, err := queryItems(s.profile)
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, resp{Msg: err.Error(), Success: false, Payload: nil})
		return
	}
	jsonResp(w, http.StatusOK, resp{Msg: "success", Success: true, Payload: titles})
}

func (s *server) LockHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if s.profile != nil {
		s.profile.Lock()
	}

	if !s.timerStartTime.IsZero() {
		// This is the recommended way to forcefully stop a timer.
		// https://golang.org/pkg/time/#Timer.Stop
		if s.timer != nil && !s.timer.Stop() {
			<-s.timer.C
		}
		s.timerStartTime = time.Time{}
	}

	// We need to be disciplined about setting this to zero everytime we stop the timer.
	jsonResp(w, http.StatusOK, resp{Msg: "success", Success: true, Payload: nil})
}

// The UnlockHandler will take the password as a payload and pass it to the configured opvault.
func (s *server) UnlockHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var payload passPayload
	if err := parsePayload(req.Body, &payload); err != nil {
		jsonResp(w, http.StatusBadRequest, resp{Msg: "bad POST payload format", Success: false, Payload: nil})
		return
	}
	if err := s.profile.Unlock(payload.Password); err != nil {
		jsonResp(w, http.StatusBadRequest, resp{Msg: "bad password", Success: false, Payload: nil})
		return
	}
	s.resetTimer()
	jsonResp(w, http.StatusOK, resp{Msg: "success", Success: true, Payload: nil})
}

func (s *server) ItemHandler(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	if s.timerStartTime.IsZero() {
		jsonResp(w, http.StatusBadRequest, resp{Msg: "vault is locked", Success: false, Payload: nil})
		return
	}
	items, err := s.profile.Items()
	if err != nil {
		jsonResp(w, http.StatusInternalServerError, resp{Msg: err.Error(), Success: false, Payload: nil})
		return
	}
	itemID := strings.ToLower(p.ByName("itemid"))
	for _, item := range items {
		if item.Trashed() || strings.ToLower(item.Title()) != itemID {
			continue
		}
		detail, err := item.Detail()
		if err != nil {
			jsonResp(w, http.StatusInternalServerError, resp{Msg: err.Error(), Success: false, Payload: nil})
			return
		}
		data := make(map[string]string)
		for _, field := range detail.Fields() {
			var name string
			if field.Designation() == "password" {
				name = "password"
			} else {
				name = field.Name()
			}
			data[name] = field.Value()
		}
		jsonResp(w, http.StatusOK, resp{Msg: "success", Success: true, Payload: data})
		return
	}
	msg := fmt.Sprintf("secret entry: \"%s\" was not found", itemID)
	jsonResp(w, http.StatusOK, resp{Msg: msg, Success: true, Payload: nil})
}
