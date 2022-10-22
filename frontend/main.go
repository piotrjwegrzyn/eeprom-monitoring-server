package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	common "pi-wegrzyn/common"
)

const version string = "2.0"

func main() {
	var configFilename = flag.String("config", "config/config.yaml", "Path to config file (YAML file)")
	var templatesDir = flag.String("templates", "templates/", "Path to templates directory (HTML files)")
	var staticDir = flag.String("static", "static/", "Path to static files (CSS and favicon)")
	var info = flag.Bool("version", false, "Print version")

	flag.Parse()

	if *info {
		fmt.Printf("Current version: %s\n", version)
		os.Exit(0)
	}

	if _, err := os.Stat(*configFilename); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Config file (%s) does not exist\n", *configFilename)
	}

	if _, err := os.Stat(*templatesDir); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Templates directory (%s) does not exist\n", *templatesDir)
	}

	if _, err := os.Stat(*staticDir); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Static directory (%s) does not exist\n", *staticDir)
	}

	log.Println("Frontend module started")

	config := common.Config{}
	common.GetConfig(*configFilename, &config)

	if err := StartServer(&config, templatesDir, staticDir); err != nil {
		log.Fatalf("Server failed with: %s\n", err)
	}
}
