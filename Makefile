VERSION = $(shell git log --pretty=format:'%h' -n 1)

all: generator

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
	./bin/eeprom-generator -c ./generator/testdata/generator.yaml -o ./bin/eeprom

.PHONY: clean
clean:
	rm -rf bin/*
