package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"pi-wegrzyn/frontend/cmds"
	"pi-wegrzyn/utils"
)

var version string

func statPaths(paths []string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	return nil
}

func main() {
	configPath := flag.String("c", "config.yaml", "Path to config file (YAML file)")
	templatesPath := flag.String("t", "templates/", "Path to templates directory (HTML files)")
	staticDir := flag.String("s", "static/", "Path to static files (CSS and favicon)")
	info := flag.Bool("v", false, "Print version")
	flag.Parse()

	if *info {
		fmt.Printf("Current version: %s\n", version)
		os.Exit(0)
	}

	if err := statPaths([]string{*configPath, *templatesPath, *staticDir}); err != nil {
		log.Fatalf("Cannot use provided paths: %v\n", err)
	}

	var cfg utils.Config
	if err := utils.ReadConfig(*configPath, &cfg); err != nil {
		log.Fatalf("Cannot read configuration: %v\n", err)
	}

	server, err := cmds.NewServer(&cfg, *templatesPath)
	if err != nil {
		log.Fatalf("Cannot prepare server: %v\n", err)
	}

	if err := server.Start(*staticDir); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}
