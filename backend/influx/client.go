package influx

import (
	"math"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type Client struct {
	config       Config
	influxClient APIWriter
}

type APIWriter interface {
	WriteAPI(org string, bucket string) api.WriteAPI
}

func New(cfg Config, influxClient APIWriter) *Client {
	return &Client{
		config:       cfg,
		influxClient: influxClient,
	}
}

type Measurement struct {
	Temperature float64
	Voltage     float64
	TxPower     float64
	RxPower     float64
	OSNR        float64
}

func (c *Client) InsertMeasurements(hostname string, interfaceName string, data Measurement) {
	writeAPI := c.influxClient.WriteAPI(c.config.Org, c.config.Bucket)

	p := influxdb2.NewPoint(
		hostname,
		map[string]string{"iface": interfaceName},
		map[string]interface{}{
			"temp":   math.Round(data.Temperature*100) / 100,
			"vcc":    math.Round(data.Voltage*100) / 100,
			"tx_pwr": math.Round(data.TxPower*100) / 100,
			"rx_pwr": math.Round(data.RxPower*100) / 100,
			"osnr":   math.Round(data.OSNR*100) / 100,
		},
		time.Now(),
	)

	writeAPI.WritePoint(p)
}
