package storage

import "fmt"

type Config struct {
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
}

func (m *Config) String() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", m.User, m.Password, m.Host, m.Port, m.Name)
}
