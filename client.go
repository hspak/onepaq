package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
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
			fmt.Println(body.Payload)
		}
	}
}

func clientAction(vars clientVars) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	switch vars.act {
	case "lock":
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
	case "unlock":
		path := fmt.Sprintf("%s/v1/1password/%s", vars.addr, vars.act)
		payload := passPayload{Password: vars.pass}
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
	case "read":
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
	default:
		msg := fmt.Sprintf("error: unknown action '%s'", vars.act)
		return errors.New(msg)
	}
	return nil
}
