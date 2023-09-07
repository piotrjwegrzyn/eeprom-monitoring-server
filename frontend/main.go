package main

import (
	"flag"
	"log"
	"os"

	"pi-wegrzyn/frontend/cmds"
	"pi-wegrzyn/utils"
)

var version string

func main() {
	configPath := flag.String("c", "config.yaml", "Path to config file (YAML file)")
	templatesPath := flag.String("t", "templates/", "Path to templates directory (HTML files)")
	staticDir := flag.String("s", "static/", "Path to static files (CSS and favicon)")
	info := flag.Bool("v", false, "Print version")
	flag.Parse()

	utils.AdjustLogger("frontend")

	if *info {
		log.Printf("Current version: %s\n", version)
		os.Exit(0)
	}

	if err := utils.StatPaths([]string{*configPath, *templatesPath, *staticDir}); err != nil {
		log.Fatalf("Cannot use provided path: %v\n", err)
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
		log.Fatalf("Server failed: %v\n", err)
	}
}
