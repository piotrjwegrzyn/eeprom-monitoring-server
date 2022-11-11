#!/bin/sh

VERSION=0.3
CONFIG_FILE=config/config.yaml

if ! $(test -f "$CONFIG_FILE");
then
    echo "Config file not found"
    exit 1
fi

USER=$(cat $CONFIG_FILE | grep -E "Users:" -A 1 | tail -1 | awk '{printf $1}' | sed 's/://')
PASSWORD=$(cat $CONFIG_FILE | grep -E "Users:" -A 1 | tail -1 | awk '{printf $2}')
DB_USER=$(cat $CONFIG_FILE | grep -E "User:" | awk '{printf $2}')
DB_PASSWORD=$(cat $CONFIG_FILE | grep -E "Passwd:" | awk '{printf $2}')
DB_NET=$(cat $CONFIG_FILE | grep -E "Net:" | awk '{printf $2}')
DB_ADDR=$(cat $CONFIG_FILE | grep -E "Addr:" | awk '{printf $2}')
DB_NAME=$(cat $CONFIG_FILE | grep -E "DBName:" | awk '{printf $2}')
PORT=$(cat $CONFIG_FILE | grep -E "Port:" | awk '{printf $2}')
BUCKET=$(cat $CONFIG_FILE | grep -E "Bucket:" | awk '{printf $2}')
ORG=$(cat $CONFIG_FILE | grep -E "Org:" | awk '{printf $2}')
TOKEN=$(cat $CONFIG_FILE | grep -E "Token:" | awk '{printf $2}')
URL=$(cat $CONFIG_FILE | grep -E "Url:" | awk '{printf $2}' | sed 's/\/\//\\\/\\\//')
RETENTION=$(cat $CONFIG_FILE | grep -E "Retention:" | awk '{printf $2}')

cd ./backend
go build
cd ../frontend
go build
cd ../

cp Dockerfile .Dockerfile.bck
sed 's/__version/'$VERSION'/' -i Dockerfile

sed 's/__config_user/'$USER'/' -i Dockerfile
sed 's/__config_password/'$PASSWORD'/' -i Dockerfile
sed 's/__config_db_user/'$DB_USER'/' -i Dockerfile
sed 's/__config_db_password/'$DB_PASSWORD'/' -i Dockerfile
sed 's/__config_db_net/'$DB_NET'/' -i Dockerfile
sed 's/__config_db_addr/'$DB_ADDR'/' -i Dockerfile
sed 's/__config_db_name/'$DB_NAME'/' -i Dockerfile
sed 's/__config_port/'$PORT'/' -i Dockerfile
sed 's/__config_bucket/'$BUCKET'/' -i Dockerfile
sed 's/__config_org/'$ORG'/' -i Dockerfile
sed 's/__config_token/'$TOKEN'/' -i Dockerfile
sed 's/__config_url/'$URL'/' -i Dockerfile
sed 's/__config_retention/'$RETENTION'/' -i Dockerfile

wget -O influx.tmp https://portal.influxdata.com/downloads/
DOWNLOAD_INFLUXD=$(cat influx.tmp | grep "wget https://dl.influxdata.com/influxdb/releases/influxdb2" | grep "linux-amd64" | awk '{printf $2 "\n"}' | head -1)
DOWNLOAD_INFLUXC=$(cat influx.tmp | grep "wget https://dl.influxdata.com/influxdb/releases/influxdb2" | grep "linux-amd64" | awk '{printf $2 "\n"}' | tail -1)
rm influx.tmp

mkdir influx_download influx
wget -O influxd.tar.gz $DOWNLOAD_INFLUXD
tar -xvf influxd.tar.gz -C ./influx_download
mv ./influx_download/$(ls influx_download | grep -v "client")/influxd ./influx/influxd
mv ./influx_download/$(ls influx_download | grep -v "client")/LICENSE ./influx/D_LICENSE
wget -O influxc.tar.gz $DOWNLOAD_INFLUXC
tar -xvf influxc.tar.gz -C ./influx_download
mv ./influx_download/$(ls influx_download | grep "client")/influx ./influx/influx
mv ./influx_download/$(ls influx_download | grep "client")/LICENSE ./influx/CLI_LICENSE

docker build -t pi-wegrzyn/eeprom-monitoring-server:$VERSION -t pi-wegrzyn/eeprom-monitoring-server:latest .

rm -rf Dockerfile influxd.tar.gz influxc.tar.gz influx_download/ influx.tmp
mv .Dockerfile.bck Dockerfile
