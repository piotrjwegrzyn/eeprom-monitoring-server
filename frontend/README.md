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

