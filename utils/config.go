package utils

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Delays struct {
	Startup int `yaml:"startup"`
	SQL     int `yaml:"sql"`
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

func StatPaths(paths []string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	return nil
}

func AdjustLogger(prefix string) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix(fmt.Sprintf("%10s: ", prefix))
}
