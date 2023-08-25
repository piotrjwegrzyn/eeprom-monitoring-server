package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"pi-wegrzyn/backend/cmds"
	"pi-wegrzyn/utils"
)

const version string = "1.2"

func main() {

	var configFilename = flag.String("config", "config.yaml", "Path to config file (YAML file)")
	var info = flag.Bool("version", false, "Print version")

	flag.Parse()

	if *info {
		fmt.Printf("Current version: %s\n", version)
		os.Exit(0)
	}

	if _, err := os.Stat(*configFilename); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Config file (%s) does not exist\n", *configFilename)
	}

	log.Println("Backend module started")

	config := utils.Config{}
	utils.GetConfig(*configFilename, &config)

	log.Printf("Startup delay set for %d seconds\n", config.Intervals.StartupDelay)
	time.Sleep(time.Duration(config.Intervals.StartupDelay) * time.Second)

	if err := cmds.StartLoop(&config); err != nil {
		log.Fatalf("Server failed with: %s\n", err)
	}
}
