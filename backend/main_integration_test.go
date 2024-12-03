//go:build integration

// To run the test properly use the following command:
// export INFLUX_BUCKET=<bucket> && \
// export INFLUX_ORG=<organization> && \
// export INFLUX_TOKEN=<token> && \
// export INFLUX_HOST=<hostname> && \
// export INFLUX_PORT=<port> && \
// go test ./. --tags=integration -cover

package main

import (
	"pi-wegrzyn/backend/influx"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/require"
)

func Test_connectToInflux(t *testing.T) {
	var influxCfg influx.Config
	if err := envconfig.Process("", &influxCfg); err != nil {
		t.Fatal("cannot read influx configuration")
	}

	_, err := connectToInfluxDB(influxCfg)

	require.NoError(t, err)
}
