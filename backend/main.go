package main

import (
	"flag"
	"log"
	"os"

	"pi-wegrzyn/backend/cmds"
	"pi-wegrzyn/utils"
)

var version string

func main() {
	var configPath = flag.String("c", "config.yaml", "Path to config file (YAML file)")
	var info = flag.Bool("v", false, "Print version")
	flag.Parse()

	utils.AdjustLogger("backend")

	if *info {
		log.Printf("Current version: %s\n", version)
		os.Exit(0)
	}

	if err := utils.StatPaths([]string{*configPath}); err != nil {
		log.Fatalf("Cannot use provided path: %v\n", err)
	}

	var cfg utils.Config
	if err := utils.ReadConfig(*configPath, &cfg); err != nil {
		log.Fatalf("Cannot read configuration: %v\n", err)
	}

	server := cmds.NewServer(&cfg)
	if err := server.Loop(); err != nil {
		log.Fatalf("Server failed with: %s\n", err)
	}
}
