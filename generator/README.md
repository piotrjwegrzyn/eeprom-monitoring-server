# eeprom-generator
Simple tool for generating some EEPROM pages of optical modules. Compatible with CMIS 5.

## Pages that are (partially) supported:
* Lower page
* Page 00h
* Page 01h
* Page 02h
* Page 04h
* Page 11h
* Page 12h
* Page 25h (OSNR only)

## Compilation
Open folder in CMD/Terminal and type:
```
go build
```
Tested on Go 1.19.

## Usage
Type in CMD/Terminal:
```
./eeprom-generator -scenario <SCENARIO_CONFIG.yaml> -modules <MODULES_CONFIG.yaml> -location <YOUR_PATH>
```

## Config creation
Sample config files are provided.
