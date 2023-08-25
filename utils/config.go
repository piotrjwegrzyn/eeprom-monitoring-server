package utils

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type DbConfig struct {
	User   string `yaml:"User"`
	Passwd string `yaml:"Passwd"`
	Net    string `yaml:"Net"`
	Addr   string `yaml:"Addr"`
	DBName string `yaml:"DBName"`
}

func (dbConfig *DbConfig) String() string {
	return fmt.Sprintf("%s:%s@%s(%s)/%s", dbConfig.User, dbConfig.Passwd, dbConfig.Net, dbConfig.Addr, dbConfig.DBName)
}

type Intervals struct {
	StartupDelay int `yaml:"StartupDelay"`
	SqlQueryInt  int `yaml:"SqlQueryInt"`
	SshQueryInt  int `yaml:"SshQueryInt"`
}

type InfluxConfig struct {
	Bucket string `yaml:"Bucket"`
	Org    string `yaml:"Org"`
	Token  string `yaml:"Token"`
	Url    string `yaml:"Url"`
}

type Config struct {
	Users     map[string]string `yaml:"Users"`
	Database  DbConfig          `yaml:"Database"`
	Port      int               `yaml:"Port"`
	Intervals Intervals         `yaml:"Intervals"`
	Influx    InfluxConfig      `yaml:"Influx"`
}

func GetConfig(filename string, configYaml *Config) {

	configFile, err := os.ReadFile(filename)

	if err != nil {
		log.Fatalf("Error while opening file %s\n", filename)
	}

	err = yaml.Unmarshal(configFile, configYaml)

	if err != nil {
		log.Fatalf("Error while parsing file %s\n", filename)
	}
}
