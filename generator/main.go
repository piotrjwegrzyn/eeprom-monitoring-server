package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"pi-wegrzyn/generator/cmds"
)

const version string = "1.2.1"

func main() {

	var scenarioFilename = flag.String("scenario", "scenario.yaml", "Location of scenario yaml config")
	var modulesFilename = flag.String("modules", "modules.yaml", "Location of modules yaml config")
	var outputPath = flag.String("location", ".", "Output location of EEPROM files")
	var info = flag.Bool("version", false, "Print version")

	flag.Parse()

	if *info {
		fmt.Printf("Current version: %s\n", version)
		os.Exit(0)
	}

	scenarioConfig := cmds.ScenarioConfig{}
	modulesConfig := cmds.ModulesConfig{}

	cmds.GetConfig(*scenarioFilename, &scenarioConfig)
	cmds.GetConfig(*modulesFilename, &modulesConfig)

	cmds.EepromToFiles(path.Join(*outputPath, "netdev1"), modulesConfig.Modules[0].Interface, cmds.CreateTimelapse(modulesConfig.Modules[0], scenarioConfig.ScenarioModules[0], scenarioConfig.Duration))
	cmds.EepromToFiles(path.Join(*outputPath, "netdev2"), modulesConfig.Modules[1].Interface, cmds.CreateTimelapse(modulesConfig.Modules[1], scenarioConfig.ScenarioModules[1], scenarioConfig.Duration))

}
