package cmds

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Module struct {
	Interface                          string  `yaml:"Interface"`
	SFF8024Identifier                  int     `yaml:"SFF8024Identifier"`
	CmisRevision                       int     `yaml:"CmisRevision"`
	ModuleRevision                     int     `yaml:"ModuleRevision"`
	MediaType                          int     `yaml:"MediaType"`
	VendorName                         string  `yaml:"VendorName"`
	DateCode                           string  `yaml:"DateCode"`
	MaxPower                           int     `yaml:"MaxPower"`
	LenghtSMF                          int     `yaml:"LenghtSMF"`
	ModuleTempMax                      float64 `yaml:"ModuleTempMax"`
	ModuleTempMin                      float64 `yaml:"ModuleTempMin"`
	TempMonHighWarningThreshold        float64 `yaml:"TempMonHighWarningThreshold"`
	TempMonLowWarningThreshold         float64 `yaml:"TempMonLowWarningThreshold"`
	VccMonHighWarningThreshold         float64 `yaml:"VccMonHighWarningThreshold"`
	VccMonLowWarningThreshold          float64 `yaml:"VccMonLowWarningThreshold"`
	OpticalPowerTxHighWarningThreshold float64 `yaml:"OpticalPowerTxHighWarningThreshold"`
	OpticalPowerTxLowWarningThreshold  float64 `yaml:"OpticalPowerTxLowWarningThreshold"`
	OpticalPowerRxHighWarningThreshold float64 `yaml:"OpticalPowerRxHighWarningThreshold"`
	OpticalPowerRxLowWarningThreshold  float64 `yaml:"OpticalPowerRxLowWarningThreshold"`
	ProgOutputPowerMin                 float64 `yaml:"ProgOutputPowerMin"`
	ProgOutputPowerMax                 float64 `yaml:"ProgOutputPowerMax"`
	GridSpacingTxx                     int     `yaml:"GridSpacingTxx"`
	CurrentLaserFrequencyTxx           int     `yaml:"CurrentLaserFrequencyTxx"`
	TargetOutputPowerTxx               float64 `yaml:"TargetOutputPowerTxx"`
}

type ModulesConfig struct {
	Modules []Module `yaml:"Modules"`
}

type Step struct {
	Endval   float64 `yaml:"endval"`
	Duration int     `yaml:"duration"`
}

type ScenarioModule struct {
	Voltage     []Step `yaml:"Voltage"`
	Temperature []Step `yaml:"Temperature"`
	TxPower     []Step `yaml:"TxPower"`
	RxPower     []Step `yaml:"RxPower"`
	Osnr        []Step `yaml:"Osnr"`
}

type ScenarioConfig struct {
	Duration        int              `yaml:"ScenarioDuration"`
	ScenarioModules []ScenarioModule `yaml:"ScenarioModules"`
}

func GetConfig[Config ModulesConfig | ScenarioConfig](filename string, configYaml *Config) {

	modulesConfig, err := os.ReadFile(filename)

	if err != nil {
		fmt.Printf("Error while opening file %s\n", filename)
		os.Exit(0)
	}

	err = yaml.Unmarshal(modulesConfig, configYaml)

	if err != nil {
		fmt.Printf("Error while parsing file %s\n", filename)
		os.Exit(0)
	}
}
