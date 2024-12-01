package utils

import (
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

type Delays struct {
	Startup float32 `yaml:"startup"`
	SQL     float32 `yaml:"sql"`
	SSH     float32 `yaml:"ssh"`
}

type Config struct {
	Users  map[string]string `yaml:"users"`
	Port   int               `yaml:"port"`
	Delays Delays            `yaml:"delays"`
	Influx Influx            `yaml:"influx"`
}

func ReadConfig(filename string, out any) error {
	cfg, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(cfg, out); err != nil {
		return err
	}

	return nil
}

func StatPaths(paths []string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	return nil
}
