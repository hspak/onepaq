package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// This value can be over-written with linker flags
// https://blog.cloudflare.com/setting-go-variables-at-compile-time/
var DevBuild = "True"

type serverVars struct {
	configPath string
	addr       string
}

type clientVars struct {
	configPath string
	act        string
	item       string
	addr       string
	pass       string
}

func main() {
	// TODO: Use a more ergonomic CLI parser
	var (
		sVars serverVars
		cVars clientVars
	)
	defaultAddr := "localhost:8080"
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverCmd.StringVar(&sVars.configPath, "config-path", "/etc/onepaq.d/onepaq.conf", "path to the config file")

	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	clientCmd.StringVar(&cVars.act, "act", "", "action to perform")
	clientCmd.StringVar(&cVars.item, "item", "", "item to take action on")
	clientCmd.StringVar(&cVars.pass, "pass", "", "password to unlock")
	clientCmd.StringVar(&cVars.addr, "addr", defaultAddr, "server to query")
	clientCmd.StringVar(&cVars.configPath, "config-path", "/etc/onepaq.d/onepaq.conf", "path to the config file")

	if len(os.Args) < 2 {
		fmt.Println("Usage of server:")
		serverCmd.PrintDefaults()
		fmt.Println()
		fmt.Println("Usage of client:")
		clientCmd.PrintDefaults()
		fmt.Println()
		os.Exit(2)
	}

	subCmd := os.Args[1]
	switch subCmd {
	case "server":
		cfg, err := parseConfig(sVars.configPath)
		if err != nil {
			log.Fatal(err)
		}
		if err := serverCmd.Parse(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		NewServer(cfg).Serve()
	case "client":
		cfg, err := parseConfig(cVars.configPath)
		if err != nil {
			log.Fatal(err)
		}
		if err := clientCmd.Parse(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		if err := clientAction(cVars, cfg); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Println("Usage of server:")
		serverCmd.PrintDefaults()
		fmt.Println()
		fmt.Println("Usage of client:")
		clientCmd.PrintDefaults()
		fmt.Println()
		os.Exit(2)
	}
}
