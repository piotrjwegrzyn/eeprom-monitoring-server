package utils

import (
	"math"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type InterfaceData struct {
	Temperature float64
	Voltage     float64
	TxPower     float64
	RxPower     float64
	Osnr        float64
}

type Influx struct {
	Bucket string `yaml:"bucket"`
	Org    string `yaml:"org"`
	Token  string `yaml:"token"`
	Url    string `yaml:"url"`
}

func (i *Influx) Insert(hostname string, iface string, data *InterfaceData) {
	client := influxdb2.NewClient(i.Url, i.Token)
	defer client.Close()
	writeAPI := client.WriteAPI(i.Org, i.Bucket)

	p := influxdb2.NewPoint(
		hostname,
		map[string]string{"iface": iface},
		map[string]interface{}{
			"temp":   math.Round(data.Temperature*100) / 100,
			"vcc":    math.Round(data.Voltage*100) / 100,
			"tx_pwr": math.Round(data.TxPower*100) / 100,
			"rx_pwr": math.Round(data.RxPower*100) / 100,
			"osnr":   math.Round(data.Osnr*100) / 100,
		},
		time.Now(),
	)

	writeAPI.WritePoint(p)
}
