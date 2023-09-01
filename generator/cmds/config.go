package cmds

type Config struct {
	Duration int      `yaml:"Duration"`
	Modules  []Module `yaml:"Modules"`
}
type Module struct {
	Interface string `yaml:"Interface"`

	// CMIS parameters
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

	Scenario Scenario `yaml:"Scenario"`
}

type Scenario struct {
	Voltage     []Step `yaml:"Voltage"`
	Temperature []Step `yaml:"Temperature"`
	TxPower     []Step `yaml:"TxPower"`
	RxPower     []Step `yaml:"RxPower"`
	Osnr        []Step `yaml:"Osnr"`
}

type Step struct {
	Endval   float64 `yaml:"endval"`
	Duration int     `yaml:"duration"`
}
