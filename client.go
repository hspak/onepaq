package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

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

func lockOutput(client *http.Client, vars clientVars) error {
	path := fmt.Sprintf("%s/v1/1password/%s", vars.addr, vars.act)
	res, err := client.Post(path, "application/json", nil)
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

func unlockOutput(client *http.Client, vars clientVars) error {
	path := fmt.Sprintf("%s/v1/1password/%s", vars.addr, vars.act)
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
	var body resp
	if err := parsePayload(res.Body, &body); err != nil {
		return err
	}
	printPayload(vars.act, body)
	return nil
}

func readOutput(client *http.Client, vars clientVars) error {
	if len(vars.item) == 0 {
		return errors.New("error: read requires an item")
	}
	queryPath := fmt.Sprintf("%s/v1/1password/item/%s", vars.addr, vars.item)
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

func clientAction(vars clientVars) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	err := error(nil)
	switch vars.act {
	case "lock":
		err = lockOutput(client, vars)
	case "unlock":
		err = unlockOutput(client, vars)
	case "read":
		err = readOutput(client, vars)
	default:
		msg := fmt.Sprintf("error: unknown action '%s'", vars.act)
		return errors.New(msg)
	}
	return err
}
