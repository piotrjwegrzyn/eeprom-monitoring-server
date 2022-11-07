# Backend module
Extracting EEPROM from network hosts and inserting to InfluxDB

## Compilation
Open folder in CMD/Terminal and type:
```
go build
```
Package `common` from main directory is needed to compile sucessfully.

Tested on Go 1.19.

## Usage
Type in CMD/Terminal:
```
./backend -config <config_file.yaml>
```

## Config file
Sample config file is provided in main `config/` directory. It contains server's startup configuration like users, MySQL and InfluxDB databases, and time steps.

Config path might be provided explicitly.

## InfluxDB configuration
To configure InfluxDB you have to download binaries for daemon and cli from [here](https://docs.influxdata.com/influxdb/v2.5/install/?t=Linux#manually-download-and-install-the-influxd-binary) and [here](https://docs.influxdata.com/influxdb/v2.5/install/?t=Linux#download-and-install-the-influx-cli).

Start daemon and configure:
```
./influxd &
./influx setup -u login -p password -t v3rY-d1ff1cUlT-t0k3n -o eeprom-monitoring-server -b sample-bucket-for-data -r 24h -f -n default --host http://:8086
```
