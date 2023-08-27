# EMS - EEPROM Monitoring Server
The server for extracting EEPROM data from SFP modules such as temperature, voltage, Tx/Rx powers an OSNR.

## Project structure

* `backend/` – main algorithm
* `common/` – package for config parsing and database interactions 
* `config/` – example configuration file and database dump
* `frontend/` – devices' configuration site
* `static/` – CSS and favicon for devices' configuration site
* `templates/` – html template files for devices' configuration site
* `build_img.sh` – script for building EMS container from scratch
* `Dockerfile` – recipe for EMS container

## Capabilities

### Configuration page

The server provides a web-based graphical interface that allows administrator to declare which network devices should be queried. Configuration consists of providing host-name, IP address, login and password or key as on picture below.

![frontend_unit.png](.github/frontend_unit.png)

### Prometheus dashboard

The configured Server periodically gain SFPs' EEPROM data from network hosts. It is stored in [Influx database](https://www.influxdata.com/). The feature of the Server is to visualize the collected data, particularly over time and in the past.

![backend_unit.png](.github/backend_unit.png)

## Pulling from Docker Hub
The container is already compiled and available on [Docker Hub](https://hub.docker.com/r/piotrjwegrzyn/eeprom-monitoring-server). To pull type in terminal:
```
docker pull piotrjwegrzyn/eeprom-monitoring-server:latest
```

## Building from source – `build_img.sh`
Adjust configuration file for your purposes and place it in `config/` folder with name `config.yaml`.
Then open folder in Terminal and type:
```
./build_img.sh
```

The `build_img.sh` script downloads Go compiler (if not available), compiles `backend` and `frontend` modules and parses Dockerfile with configuration provided via `config.yaml`.
Also InfluxDB is downloaded.

NOTE: Modules should be compiled with Go 1.19 or later.

## Usage
Then, the Docker container is ready and you can start it typing:
```
docker run [-ti/d] [--rm] -p 80:<CONFIG_PORT> -p 8086:8086 piotrjwegrzyn/eeprom-monitoring-server:latest
```

Flags `-ti` or `-d` will determine if container will be started in "Terminal Interaction" mode or "Detached".

By default, CONFIG_PORT is 80 and there is configuration page available. Prometheus dashboard is on port 8086.

## Usage in GNS3

This container can be used in GNS3 simulation. To import it download [template file](.github/ems-template.gns3a) and import to GNS3.

![gns3.png](.github/gns3.png)


## Config file
Sample config file is provided in main `config/` directory. It contains server's startup configuration like users, MySQL and InfluxDB databases, and time steps.

Config path might be provided explicitly.

## License
[GNU GENERAL PUBLIC LICENSE](https://github.com/piotrjwegrzyn/eeprom-monitoring-server/blob/master/LICENSE)

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

# Frontend module
Configuration portal for devices

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
./frontend -config <config_file.yaml> -templates <path/to/templates-dir> -static <path/to/static-dir>
```

## Config and Templates
Sample files are provided in main directory:
* `config/` - contains server's startup configuration (users and database connection)
* `templates/` - contains HTML files
* `static/` - contains CSS and favicon files

Unlike other files config path might be provided explicitly.

Sample database file is attached in `config/` folder (tested on MariaDB 10.9.2).

# eeprom-presenter
Simple container with commands to list optical interfaces and show EEPROM on them.

## How to build on any Linux
Prepare EEPROM files for interfaces and move them to `eeproms` (remove default `eth0` folder before).

**Note**: folders' names in `eeproms` directory should reflect network interfaces' names (what is not cover here).

Then:
```
docker build -t pi-wegrzyn/eeprom-presenter:latest .
```

## How to run on any Linux
```
docker run -ti pi-wegrzyn/eeprom-presenter
```

## Usage
```
# Show optic ports:
show-fiber-interfaces

# Show EEPROM on interface:
show-eeprom <IFACE>
```

## How to run in GNS3's project
SSH to our GNS3 server and build container as above. Then follow [that guide](https://docs.gns3.com/docs/emulators/docker-support-in-gns3).

**Note**: Default GNS3 interfaces are enumerating as eth0, eth1 and so on.

**Note**: Do not type `exit` in console because it leads to shutdown container.

## License
All software is under GNU GPL version 3.

# Draft: New README
## eeprom-generator
Simple tool for generating some EEPROM pages of optical modules based on scenario predefined. Compatible with CMIS 5. Pages that are (partially) supported:
* Lower page
* Page 00h
* Page 01h
* Page 02h
* Page 04h
* Page 11h
* Page 12h
* Page 25h (OSNR only)

### Usage
Type in CMD/Terminal:
```
./eeprom-generator -c <CONFIG_FILE.yaml> -o <OUTPUT_PATH>
```

### Config file
Sample config file is provided in `testdata/generator.yaml`. 
In configuration file there is a scenario's duration defined in seconds and modules list. The `Modules` list contains an interface name, CMIS content data and a `Scenario` with a bunch of fiber-working parameters as below:

```
  Scenario:
    Temperature:
      - endval: 33.0
        duration: 120
      - endval: 32.0
        duration: 180
```

For the first 120 seconds the value of `Temperature` will be interpreted as `33.0` like a linear function with start and end points (X, Y) set as (1, 33.0) and (33.0, 120). Then from the 121st second till the end (120+180) the `Temperature` will slowly decrease like a linear function with starting point (X, Y) as (121, 33.0) and ending point as (300, 32.0).

If there is a need of preparing an instant change we are able to change that defining a one-second-event as the middle step below:
```
  Scenario:
    Temperature:
      - endval: 33.0
        duration: 120
      - endval: 10.0
        duration: 1
      - endval: 10.0
        duration: 179
```

**Note**: the very first step will always be a flat function with defined `endval`. If you want to start linear change from the beginning you should create one-second-event step.
