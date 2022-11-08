#!/bin/sh

VERSION=0.1

cd ./backend
go build
cd ../frontend
go build
cd ../

## TODO:
## sed for Dockerfile (ARGs)

sed 's/__version/'$VERSION'/' -i Dockerfile

docker build -t pi-wegrzyn/eeprom-monitoring-server:$VERSION -t pi-wegrzyn/eeprom-monitoring-server:latest .
