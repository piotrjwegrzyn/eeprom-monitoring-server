package utils

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Delays struct {
	Startup int `yaml:"startup"`
	Query   int `yaml:"query"`
	SSH     int `yaml:"ssh"`
}

type Config struct {
	Users  map[string]string `yaml:"users"`
	MySQL  MySQL             `yaml:"mysql"`
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
