package cmds

import (
	"fmt"
)

func generateSteps(steps []Step, duration int) (listS []float64) {
	for i := 0; i < len(steps); i++ {
		if i == 0 {
			for j := 0; j < steps[i].Duration; j++ {
				listS = append(listS, steps[i].Endval)
			}
		} else {
			f := func(x float64) float64 {
				return (steps[i].Endval-steps[i-1].Endval)*(x+1)/float64(steps[i].Duration) + steps[i-1].Endval
			}
			for j := 0; j < steps[i].Duration; j++ {
				listS = append(listS, f(float64(j)))
			}
		}
	}

	if len(listS) != duration {
		fmt.Printf("Error: mismatch between ScenarioDuration (%d) and steps durations (%d) near %f endval\n", duration, len(listS), listS[len(listS)-1])
		panic(1)
	}

	return
}

func CreateTimelapse(module Module, scenario ScenarioModule, duration int) (timelapse [][]byte) {

	listVcc := generateSteps(scenario.Voltage, duration)
	listTemp := generateSteps(scenario.Temperature, duration)
	listTxPower := generateSteps(scenario.TxPower, duration)
	listRxPower := generateSteps(scenario.RxPower, duration)
	listOsnr := generateSteps(scenario.Osnr, duration)

	page00h := GeneratePage00h(module)
	page01h := GeneratePage01h(module)
	page02h := GeneratePage02h(module)
	page04h := GeneratePage04h(module)
	page12h := GeneratePage12h(module)

	for i := 0; i < duration; i++ {
		timelapseStep := make([]byte, 0)
		timelapseStep = append(timelapseStep, GeneratePageLow(module, listTemp[i], listVcc[i])...)
		timelapseStep = append(timelapseStep, page00h...)
		timelapseStep = append(timelapseStep, page01h...)
		timelapseStep = append(timelapseStep, page02h...)
		timelapseStep = append(timelapseStep, page04h...)
		timelapseStep = append(timelapseStep, GeneratePage11h(module, listTxPower[i], listRxPower[i])...)
		timelapseStep = append(timelapseStep, page12h...)
		timelapseStep = append(timelapseStep, GeneratePage25h(listOsnr[i], listTemp[i])...)
		timelapse = append(timelapse, timelapseStep)
	}

	return
}
