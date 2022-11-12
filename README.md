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

## Building – `build_img.sh`
Adjust configuration file for your purposes and place it in `config/` folder with name `config.yaml`.
Then open folder in Terminal and type:
```
./build_img.sh
```

The `build_img.sh` script downloads Go compiler (if not available), compiles `backend` and `frontend` modules and parses Dockerfile with configuration provided via `config.yaml`.
Also InfluxDB is downloaded.

NOTE: Modules should be compiled with Go 1.19 or later.

## Usage
Then, the Docker container is built and you can start it typing:
```
docker run [-ti/d] [--rm] pi-wegrzyn/eeprom-monitoring-server:latest
```

Flags `-ti` or `-d` will determine if container will be started in "Terminal Interaction" mode or "Detached".

## Config file
Sample config file is provided in main `config/` directory. It contains server's startup configuration like users, MySQL and InfluxDB databases, and time steps.

Config path might be provided explicitly.
