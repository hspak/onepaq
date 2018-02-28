package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

// TODO: this entire client action handling can be cleaned up.

func printPayload(act string, body resp) {
	if !body.Success {
		fmt.Printf("Failed to %s: %s\n", act, body.Msg)
	} else if body.Payload == nil {
		if act == "read" {
			fmt.Println("not found")
		} else {
			fmt.Println("success")
		}
	} else {
		payload := body.Payload.(map[string]interface{})
		if val, ok := payload["password"]; ok {
			fmt.Println(val)
		} else {
			fmt.Println("could not find password, printing entire payload")
			fmt.Println(body.Payload)
		}
	}
}

func lockOutput(proto string, client *http.Client, vars clientVars) error {
	path := fmt.Sprintf("%s://%s/v1/1password/%s", proto, vars.addr, vars.act)
	res, err := client.Post(path, "application/json", nil)
	if err != nil {
		return err
	}
	var body resp
	if err := parsePayload(res.Body, &body); err != nil {
		return err
	}
	defer res.Body.Close()
	printPayload(vars.act, body)
	return nil
}

func unlockOutput(proto string, client *http.Client, vars clientVars) error {
	path := fmt.Sprintf("%s://%s/v1/1password/%s", proto, vars.addr, vars.act)
	var pass string
	if len(vars.pass) == 0 {
		fmt.Printf("Vault password: ")
		inputPass, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Println()
		pass = string(inputPass)
	} else {
		pass = vars.pass
	}
	payload := passPayload{Password: pass}
	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	res, err := client.Post(path, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var body resp
	if err := parsePayload(res.Body, &body); err != nil {
		return err
	}
	printPayload(vars.act, body)
	return nil
}

func readOutput(proto string, client *http.Client, vars clientVars) error {
	if len(vars.item) == 0 {
		return errors.New("error: read requires an item")
	}
	queryPath := fmt.Sprintf("%s://%s/v1/1password/item/%s", proto, vars.addr, vars.item)
	res, err := client.Get(queryPath)
	if err != nil {
		return err
	}

	var body resp
	if err := parsePayload(res.Body, &body); err != nil {
		return err
	}
	printPayload(vars.act, body)
	return nil
}

func setupTLSConfig(cfg config) (*tls.Config, error) {
	tlsClientConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		tlsCert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, err
		}
		tlsClientConfig.Certificates = []tls.Certificate{tlsCert}
	}

	if cfg.CertCA != "" {
		caCert, err := ioutil.ReadFile(cfg.CertCA)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsClientConfig.RootCAs = caCertPool
	}

	return tlsClientConfig, nil
}

func clientAction(vars clientVars, cfg config) error {
	tlsConfig, err := setupTLSConfig(cfg)
	if err != nil {
		return err
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	var proto string
	if cfg.CertFile != "" && cfg.KeyFile != "" && cfg.CertCA != "" {
		proto = "https"
	} else {
		proto = "http"
	}

	err = error(nil)
	switch vars.act {
	case "lock":
		err = lockOutput(proto, client, vars)
	case "unlock":
		err = unlockOutput(proto, client, vars)
	case "read":
		err = readOutput(proto, client, vars)
	default:
		msg := fmt.Sprintf("error: unknown action '%s'", vars.act)
		return errors.New(msg)
	}
	return err
}
