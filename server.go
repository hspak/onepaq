package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hspak/opvault"
	"github.com/julienschmidt/httprouter"
)

// Different data sources would go here
type server struct {
	profile    *opvault.Profile
	timer      *time.Timer
	timeout    time.Duration // seconds
	timerStart time.Time
	port       string
	logfile    *os.File
}

func NewServer(cfg config) *server {
	server := new(server)
	vault, err := opvault.Open(cfg.OpvaultPath)
	if err != nil {
		log.Fatal(err)
	}
	if err = checkProfiles(vault, cfg.ProfileName); err != nil {
		log.Fatal(err)
	}
	profile, err := vault.Profile(cfg.ProfileName)
	if err != nil {
		log.Fatal(err)
	}
	server.logfile, err = setupLogging(cfg.LogPath)
	if err != nil {
		log.Fatal(err)
	}

	// This will be initialized on first use.
	server.timer = nil

	server.timeout = cfg.UnlockTimeout * time.Second
	server.timerStart = time.Time{}
	server.profile = profile
	server.port = strconv.Itoa(cfg.HTTPPort)
	return server
}

func (s *server) resetTimer() {
	if s.timer == nil {
		s.timer = time.NewTimer(s.timeout)
		s.log("DEBUG", fmt.Sprintf("timer first initialized with timeout %s", s.timeout))
	} else {
		s.timer.Reset(s.timeout)
		remaining := s.timeout - time.Since(s.timerStart)
		s.log("DEBUG", fmt.Sprintf("timer reset with timeout %s (had %s remaining)", s.timeout, remaining))
	}
	s.timerStart = time.Now()
	go func() {
		<-s.timer.C
		// We need to be disciplined about setting this to zero everytime we stop the timer.
		s.timerStart = time.Time{}
		s.profile.Lock()
		s.log("INFO", "vault unlock timed out, vault is locked")
	}()
}

// This is discount logging with fake levels.
// TODO: pull in a real logger.
func (s *server) log(level, msg string) {
	time := time.Now().Format(time.RFC3339)
	fullMsg := fmt.Sprintf("%s - [%s] - %s\n", time, level, msg)
	fmt.Fprintf(s.logfile, "%s", fullMsg)
}

func setupLogging(logPath string) (*os.File, error) {
	if DevBuild == "True" {
		return os.Stdout, nil
	}
	logfile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return logfile, nil
}

func parsePayload(payload io.ReadCloser, obj interface{}) error {
	// This loads the entire payload into memory.
	body, err := ioutil.ReadAll(payload)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, obj)
	if err != nil {
		return err
	}
	return nil
}

func (s *server) Serve() {
	mux := httprouter.New()
	// TODO: setup a frontend, eventually
	mux.GET("/v1/1password/status", s.StatusHandler)
	mux.GET("/v1/1password/items", s.ItemsHandler)
	// mux.GET("/v1/1password/item/:itemid", s.ItemsHandler)
	mux.GET("/v1/1password/item/:itemid/password", s.PasswordHandler)
	mux.POST("/v1/1password/lock", s.LockHandler)
	mux.POST("/v1/1password/unlock", s.UnlockHandler)
	s.log("INFO", fmt.Sprintf("listening on port %s", s.port))
	log.Fatal(http.ListenAndServe(":"+s.port, mux))
}
