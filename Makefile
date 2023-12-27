VERSION = latest-$(shell git log --pretty=format:'%h' -n 1)

all: ems generator

.PHONY: backend
backend:
	go build -C backend -o ../bin/ems-backend -ldflags "-X main.version=$(VERSION)"

.PHONY: frontend
frontend:
	go build -C frontend -o ../bin/ems-frontend -ldflags "-X main.version=$(VERSION)"

.PHONY: influxdb
influxdb:
	$(eval LATEST=$(shell wget -q -O- https://www.influxdata.com/downloads/ | grep ">wget https://dl.influxdata.com/influxdb/releases/influxdb2" | grep "linux" | grep "amd64" | awk '{printf $$2 "\n"}'))
	$(eval INFLUX_VERSION=$(shell echo $(LATEST) | awk '{printf $$1}' | grep -o -E "[0-9]+\.[0-9]+\.[0-9]+"))

	wget -q -O - $(shell echo $(LATEST) | awk '{printf $$1}') | tar -zxv influxdb2-$(shell echo $(INFLUX_VERSION))/usr/bin/influxd --transform 's/influxdb2-$(shell echo $(INFLUX_VERSION))\/usr\/bin\/influxd/bin\/influxd/'
	wget -q -O - $(shell echo $(LATEST) | awk '{printf $$2}') | tar -zxv ./influx --transform 's/\.\/influx/bin\/influxc/'

.PHONY: ems
ems: backend frontend influxdb ems-build

.PHONY: ems-build
ems-build:
	docker build --no-cache \
	--file Dockerfile \
	--tag pi-wegrzyn/ems:$(VERSION) \
	--tag pi-wegrzyn/ems:latest \
	--build-arg CONFIG=./testdata/ems.yaml .

.PHONY: ems-cached
ems-build-cached:
	docker build \
	--file Dockerfile \
	--tag pi-wegrzyn/ems:$(VERSION) \
	--tag pi-wegrzyn/ems:latest \
	--build-arg CONFIG=./testdata/ems.yaml .

.PHONY: generator
generator:
	go build -C generator -o ../bin/eeprom-generator -ldflags "-X main.version=$(VERSION)"

.PHONY: presenter
presenter:
	docker build --no-cache \
	--file presenter/Dockerfile \
	--tag pi-wegrzyn/ep:$(VERSION) \
	--tag pi-wegrzyn/ep:latest \
	--build-arg EEPROM_ITER=300 \
	--build-arg SLEEP_TIME=1 \
	--build-arg EEPROM_SRC=bin/eeprom/ .

.PHONY: sample-presenter
sample-presenter: generator sample-eeprom-files presenter

.PHONY: sample-eeprom-files
sample-eeprom-files:
	./bin/eeprom-generator -c ./testdata/generator.yaml -o ./bin/eeprom

.PHONY: clean
clean:
	rm -rf bin/*
