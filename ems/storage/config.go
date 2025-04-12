package storage

import "fmt"

type Config struct {
	Name     string `envconfig:"DB_NAME"`
	User     string `envconfig:"DB_USER"`
	Password string `envconfig:"DB_PASSWORD"`
	Host     string `envconfig:"DB_HOST"`
	Port     string `envconfig:"DB_PORT"`
}

func (m *Config) String() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", m.User, m.Password, m.Host, m.Port, m.Name)
}
