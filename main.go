package main

import (
	"flag"
	"log"
)

// This value can be over-written with linker flags
// https://blog.cloudflare.com/setting-go-variables-at-compile-time/
var DevBuild = "True"

func main() {
	var configPath string
	flag.StringVar(&configPath, "config-path", "/etc/onepaq.conf", "path to the config file")
	flag.Parse()
	config, err := parseConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	NewServer(config).Serve()
}
