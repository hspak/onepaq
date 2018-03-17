package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hspak/opvault"
	"github.com/julienschmidt/httprouter"
)

type server struct {
	profile        *opvault.Profile
	timer          *time.Timer
	timeout        time.Duration // seconds
	timerStartTime time.Time
	addr           string
	logfile        *os.File
	tlsCert        string
	tlsKey         string
	tlsRootCA      *x509.CertPool
}

func NewServer(cfg config) *server {
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
	logfile, err := setupLogging("/var/log/onepaq/onepaq.log")
	if err != nil {
		log.Fatal(err)
	}
	var rootCAs *x509.CertPool
	if cfg.CertCA != "" {
		caCert, err := ioutil.ReadFile(cfg.CertCA)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		rootCAs = caCertPool
	}

	return &server{
		logfile:        logfile,
		timer:          nil, // This will be initialized on first use.
		tlsKey:         cfg.KeyFile,
		tlsCert:        cfg.CertFile,
		tlsRootCA:      rootCAs,
		timeout:        cfg.UnlockTimeout * time.Second,
		timerStartTime: time.Time{},
		profile:        profile,
		addr:           cfg.HTTPAddr,
	}
}

func (s *server) resetTimer() {
	if s.timer == nil {
		s.timer = time.NewTimer(s.timeout)
		s.log("DEBUG", fmt.Sprintf("vault lock timer first triggered with timeout %s", s.timeout))
	} else {
		if !s.timerStartTime.IsZero() {
			remaining := s.timeout - time.Since(s.timerStartTime)
			s.log("DEBUG", fmt.Sprintf("vault lock timer reset manually (had %s remaining)", remaining))
		} else {
			s.log("DEBUG", fmt.Sprintf("vault lock timer reset with timeout %s", s.timeout))
		}
		s.timer.Reset(s.timeout)
	}
	s.timerStartTime = time.Now()
	go func() {
		<-s.timer.C
		// We need to be disciplined about setting this to zero everytime we stop the timer.
		s.timerStartTime = time.Time{}
		s.profile.Lock()
		s.log("INFO", "vault lock timer ran out, vault is locked")
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
	logfile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return logfile, nil
}

func parsePayload(payload io.Reader, obj interface{}) error {
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
	mux.GET("/v1/1password/status", s.StatusHandler)
	mux.GET("/v1/1password/items", s.ItemsHandler)
	mux.GET("/v1/1password/item/:itemid", s.ItemHandler)
	mux.POST("/v1/1password/lock", s.LockHandler)
	mux.POST("/v1/1password/unlock", s.UnlockHandler)
	if s.tlsKey != "" && s.tlsCert != "" {
		s.log("INFO", fmt.Sprintf("listening on %s with TLS", s.addr))
		tlsCfg := &tls.Config{
			ClientAuth:         tls.RequireAndVerifyClientCert,
			InsecureSkipVerify: false,
			ClientCAs:          s.tlsRootCA,
		}
		srv := http.Server{
			Addr:              s.addr,
			Handler:           mux,
			TLSConfig:         tlsCfg,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      5 * time.Second,
		}
		log.Fatal(srv.ListenAndServeTLS(s.tlsCert, s.tlsKey))
	} else {
		s.log("INFO", fmt.Sprintf("listening on %s", s.addr))
		log.Fatal(http.ListenAndServe(s.addr, mux))
	}
}
