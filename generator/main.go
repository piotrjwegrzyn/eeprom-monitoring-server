package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"pi-wegrzyn/generator/cmds"
	"pi-wegrzyn/utils"
)

var version string

func main() {
	configPath := flag.String("c", "config.yaml", "Path to configuration file")
	outputPath := flag.String("o", ".", "Output location of EEPROM files")
	info := flag.Bool("v", false, "Print version")
	flag.Parse()

	if *info {
		fmt.Printf("Current version: %s\n", version)
		os.Exit(0)
	}

	var cfg cmds.Config
	if err := utils.ReadConfig(*configPath, &cfg); err != nil {
		log.Fatalf("Cannot read configuration: %v\n", err)
	}

	for i := 0; i < len(cfg.Modules); i++ {
		out := path.Join(*outputPath, cfg.Modules[i].Interface)
		timelapse, err := cmds.CreateTimelapse(cfg.Modules[i], cfg.Duration)
		if err != nil {
			log.Fatalf("Cannot generate timelapse: %v", err)
		}

		if err := cmds.SaveToFile(out, cfg.Modules[i].Interface, timelapse); err != nil {
			log.Fatalf("Cannot save file: %v", err)
		}
	}
}
