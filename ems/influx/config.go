package influx

type Config struct {
	Bucket string `envconfig:"BUCKET"`
	Org    string `envconfig:"ORG"`
	Token  string `envconfig:"TOKEN"`
	Host   string `envconfig:"HOST"`
	Port   string `envconfig:"PORT"`
}
