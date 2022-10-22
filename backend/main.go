package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	common "pi-wegrzyn/common"
	"time"
)

const version string = "0.1-prealpha"

func main() {

	var configFilename = flag.String("config", "config/config.yaml", "Path to config file (YAML file)")
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

	config := common.Config{}
	common.GetConfig(*configFilename, &config)

	log.Printf("Delay set for %d seconds\n", config.Intervals.StartupDelay)

	time.Sleep(time.Duration(config.Intervals.StartupDelay) * time.Second)

}
