all: backend frontend generator presenter influxdb

.PHONY: backend
backend:
	go build -C backend -o ../bin/ems-backend

.PHONY: frontend
frontend:
	go build -C frontend -o ../bin/ems-frontend

.PHONY: generator
generator:
	go build -C generator -o ../bin/eeprom-generator

.PHONY: presenter
presenter:
	docker build --file presenter/Dockerfile --tag pi-wegrzyn/ep:latest presenter

.PHONY: influxdb
influxdb:
	$(eval LATEST=$(shell wget -q -O- https://portal.influxdata.com/downloads/ | grep "wget https://dl.influxdata.com/influxdb/releases/influxdb2" | grep "linux-amd64" | awk '{printf $$2 "\n"}'))

	wget -q -O - $(shell echo $(LATEST) | awk '{printf $$1}') | tar -zxv influxdb2_linux_amd64/influxd --transform 's/influxdb2_linux_amd64\/influxd/bin\/influxd/'
	wget -q -O - $(shell echo $(LATEST) | awk '{printf $$2}') | tar -zxv ./influx --transform 's/\.\/influx/bin\/influxc/'

.PHONY: clean
clean:
	rm -rf bin/*