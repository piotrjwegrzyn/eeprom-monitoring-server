//go:build integration

// To run the test properly use the following command:
// export INFLUX_BUCKET=<bucket> && \
// export INFLUX_ORG=<organization> && \
// export INFLUX_TOKEN=<token> && \
// export INFLUX_HOST=<hostname> && \
// export INFLUX_PORT=<port> && \
// go test ./... --tags=integration -cover

package influx

import (
	"context"
	"fmt"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/kelseyhightower/envconfig"
)

func connect() (*influxdb2.Client, error) {
	var cfg struct {
		Bucket string `envconfig:"INFLUX_BUCKET" required:"true" validate:"required"`
		Org    string `envconfig:"INFLUX_ORG" required:"true" validate:"required"`
		Token  string `envconfig:"INFLUX_TOKEN" required:"true" validate:"required"`
		Host   string `envconfig:"INFLUX_HOST" required:"true" validate:"required"`
		Port   string `envconfig:"INFLUX_PORT" required:"true" validate:"required"`
	}

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	client := influxdb2.NewClient(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), cfg.Token)
	_, err := client.Health(context.Background())
	return &client, err
}

func TestIntegration_InsertMeasurements(t *testing.T) {
	// TODO
}
