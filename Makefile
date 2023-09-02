VERSION = latest-$(shell git log --pretty=format:'%h' -n 1)

all: backend frontend influxdb ems generator eeprom presenter

.PHONY: backend
backend:
	go build -C backend -o ../bin/ems-backend -ldflags "-X main.version=$(VERSION)"

.PHONY: frontend
frontend:
	go build -C frontend -o ../bin/ems-frontend -ldflags "-X main.version=$(VERSION)"

.PHONY: influxdb
influxdb:
	$(eval LATEST=$(shell wget -q -O- https://portal.influxdata.com/downloads/ | grep "wget https://dl.influxdata.com/influxdb/releases/influxdb2" | grep "linux-amd64" | awk '{printf $$2 "\n"}'))

	wget -q -O - $(shell echo $(LATEST) | awk '{printf $$1}') | tar -zxv influxdb2_linux_amd64/influxd --transform 's/influxdb2_linux_amd64\/influxd/bin\/influxd/'
	wget -q -O - $(shell echo $(LATEST) | awk '{printf $$2}') | tar -zxv ./influx --transform 's/\.\/influx/bin\/influxc/'

.PHONY: ems
ems:
	docker build --no-cache \
	--file Dockerfile \
	--tag pi-wegrzyn/ems:$(VERSION) .

.PHONY: generator
generator:
	go build -C generator -o ../bin/eeprom-generator -ldflags "-X main.version=$(VERSION)"

.PHONY: eeprom
eeprom:
	./bin/eeprom-generator -c ./testdata/generator.yaml -o ./bin/eeprom

.PHONY: presenter
presenter:
	docker build --no-cache \
	--file presenter/Dockerfile \
	--tag pi-wegrzyn/ep:$(VERSION) \
	--build-arg EEPROM_ITER=300 \
	--build-arg SLEEP_TIME=1 \
	--build-arg EEPROM_SRC=bin/eeprom/ .

.PHONY: clean
clean:
	rm -rf bin/*
