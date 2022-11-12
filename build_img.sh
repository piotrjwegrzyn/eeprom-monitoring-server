#!/bin/sh

VERSION=0.4
CONFIG_FILE=config/config.yaml

if ! $(test -f "$CONFIG_FILE");
then
    echo "Config file not found"
    exit 1
fi

cd ./backend
go build
cd ../frontend
go build
cd ../

cp Dockerfile .Dockerfile.bck
sed 's/__version/'$VERSION'/' -i Dockerfile

sed 's/__config_user/'$(cat $CONFIG_FILE | grep -E "Users:" -A 1 | tail -1 | awk '{printf $1}' | sed 's/://')'/' -i Dockerfile
sed 's/__config_password/'$(cat $CONFIG_FILE | grep -E "Users:" -A 1 | tail -1 | awk '{printf $2}')'/' -i Dockerfile
sed 's/__config_db_user/'$(cat $CONFIG_FILE | grep -E "User:" | awk '{printf $2}')'/' -i Dockerfile
sed 's/__config_db_password/'$(cat $CONFIG_FILE | grep -E "Passwd:" | awk '{printf $2}')'/' -i Dockerfile
sed 's/__config_db_name/'$(cat $CONFIG_FILE | grep -E "DBName:" | awk '{printf $2}')'/' -i Dockerfile
sed 's/__config_port/'$(cat $CONFIG_FILE | grep -E "Port:" | awk '{printf $2}')'/' -i Dockerfile
sed 's/__config_bucket/'$(cat $CONFIG_FILE | grep -E "Bucket:" | awk '{printf $2}')'/' -i Dockerfile
sed 's/__config_org/'$(cat $CONFIG_FILE | grep -E "Org:" | awk '{printf $2}')'/' -i Dockerfile
sed 's/__config_token/'$(cat $CONFIG_FILE | grep -E "Token:" | awk '{printf $2}')'/' -i Dockerfile
sed 's/__config_retention/'$(cat $CONFIG_FILE | grep -E "Retention:" | awk '{printf $2}')'/' -i Dockerfile

wget -O influx.tmp https://portal.influxdata.com/downloads/
DOWNLOAD_INFLUXD=$(cat influx.tmp | grep "wget https://dl.influxdata.com/influxdb/releases/influxdb2" | grep "linux-amd64" | awk '{printf $2 "\n"}' | head -1)
DOWNLOAD_INFLUXC=$(cat influx.tmp | grep "wget https://dl.influxdata.com/influxdb/releases/influxdb2" | grep "linux-amd64" | awk '{printf $2 "\n"}' | tail -1)

mkdir influx_download influx

if ! $(test -f "./influx/influxd");
then
    echo "Downloading Influx daemon..."
    wget -O influxd.tar.gz $DOWNLOAD_INFLUXD
    tar -xvf influxd.tar.gz -C ./influx_download
    mv ./influx_download/$(ls influx_download | grep -v "client")/influxd ./influx/influxd
    mv ./influx_download/$(ls influx_download | grep -v "client")/LICENSE ./influx/D_LICENSE
fi

if ! $(test -f "./influx/influx");
then
    echo "Downloading Influx CLI..."
    wget -O influxc.tar.gz $DOWNLOAD_INFLUXC
    tar -xvf influxc.tar.gz -C ./influx_download
    mv ./influx_download/$(ls influx_download | grep "client")/influx ./influx/influx
    mv ./influx_download/$(ls influx_download | grep "client")/LICENSE ./influx/CLI_LICENSE
fi

docker build -t pi-wegrzyn/eeprom-monitoring-server:$VERSION -t pi-wegrzyn/eeprom-monitoring-server:latest .

rm -rf influxd.tar.gz influxc.tar.gz influx_download/ influx.tmp
mv Dockerfile Dockerfile_v"$VERSION"_$(date +%F_%s)
mv .Dockerfile.bck Dockerfile
