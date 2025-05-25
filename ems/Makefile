VERSION = $(shell git log --pretty=format:'%h' -n 1)

all: ems

.PHONY: influxdb
influxdb:
	$(eval LATEST_DAEMON=$(shell wget -q -O- https://www.influxdata.com/downloads/ | grep ">wget https://download.influxdata.com/influxdb/releases/influxdb2" | grep "linux" | grep "amd64" | awk '{printf $$2 "\n"}'))
	$(eval LATEST_CLIENT=$(shell wget -q -O- https://www.influxdata.com/downloads/ | grep ">wget https://dl.influxdata.com/influxdb/releases/influxdb2-client" | grep "linux" | grep "amd64" | awk '{printf $$2 "\n"}'))
	$(eval INFLUXD_VERSION=$(shell echo $(LATEST_DAEMON) | awk '{printf $$1}' | grep -o -E "[0-9]+\.[0-9]+\.[0-9]+"))

	wget -q -O - $(shell echo $(LATEST_DAEMON) | awk '{printf $$1}') | tar -zxv influxdb2-$(shell echo $(INFLUXD_VERSION))/usr/bin/influxd --transform 's/influxdb2-$(shell echo $(INFLUXD_VERSION))\/usr\/bin\/influxd/bin\/influxd/'
	wget -q -O - $(shell echo $(LATEST_CLIENT) | awk '{printf $$1}') | tar -zxv ./influx --transform 's/\.\/influx/bin\/influxc/'

.PHONY: ems
ems: influxdb ems-build

.PHONY: ems-build
ems-build:
	docker build --no-cache \
	--file Dockerfile \
	--target aio \
	--tag pi-wegrzyn/ems:$(VERSION) \
	--tag pi-wegrzyn/ems:latest .

.PHONY: ems-build-cached
ems-build-cached:
	docker build \
	--file Dockerfile \
	--target aio \
	--tag pi-wegrzyn/ems:$(VERSION) \
	--tag pi-wegrzyn/ems:latest .

.PHONY: clean
clean:
	rm -rf bin/*
