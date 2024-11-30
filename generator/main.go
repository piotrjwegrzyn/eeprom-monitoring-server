package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"

	"pi-wegrzyn/generator/cmds"
	"pi-wegrzyn/utils"
)

type ctxKey string

const appNameAttr ctxKey = "app"

var version string

func main() {
	appCtx := context.WithValue(context.Background(), appNameAttr, "generator")

	configPath := flag.String("c", "config.yaml", "Path to configuration file")
	outputPath := flag.String("o", ".", "Output location of EEPROM files")
	info := flag.Bool("v", false, "Print version")
	flag.Parse()

	if *info {
		slog.InfoContext(appCtx, fmt.Sprintf("current version: %s", version))
		os.Exit(0)
	}

	var cfg cmds.Config
	if err := utils.ReadConfig(*configPath, &cfg); err != nil {
		slog.ErrorContext(appCtx, "cannot read configuration", slog.Any("error", err))
		os.Exit(1)
	}

	for i := 0; i < len(cfg.Modules); i++ {
		out := path.Join(*outputPath, cfg.Modules[i].Interface)
		timelapse, err := cmds.CreateTimelapse(cfg.Modules[i], cfg.Duration)
		if err != nil {
			slog.ErrorContext(appCtx, "cannot generate timelapse", slog.Any("error", err))
			os.Exit(1)
		}

		if err := cmds.SaveToFile(out, cfg.Modules[i].Interface, timelapse); err != nil {
			slog.ErrorContext(appCtx, "cannot save file", slog.Any("error", err))
			os.Exit(1)
		}
	}
}
