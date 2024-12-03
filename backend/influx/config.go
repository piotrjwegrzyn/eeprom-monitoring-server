package influx

type Config struct {
	Bucket string `envconfig:"INFLUX_BUCKET"`
	Org    string `envconfig:"INFLUX_ORG"`
	Token  string `envconfig:"INFLUX_TOKEN"`
	Host   string `envconfig:"INFLUX_HOST"`
	Port   string `envconfig:"INFLUX_PORT"`
}
