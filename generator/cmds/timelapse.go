package cmds

import (
	"fmt"
	"os"
	"path"
)

func generateSteps(steps []Step, duration int) ([]float64, error) {
	newSteps := make([]float64, 0, duration)

	for i := 0; i < len(steps); i++ {
		if i == 0 {
			for j := 0; j < steps[i].Duration; j++ {
				newSteps = append(newSteps, steps[i].Endval)
			}
		} else {
			f := func(x float64) float64 {
				return (steps[i].Endval-steps[i-1].Endval)*(x+1)/float64(steps[i].Duration) + steps[i-1].Endval
			}
			for j := 0; j < steps[i].Duration; j++ {
				newSteps = append(newSteps, f(float64(j)))
			}
		}
	}

	if len(newSteps) != duration {
		return nil, fmt.Errorf("mismatch between Duration (%d) and steps durations (%d) near %f endval\n", duration, len(newSteps), newSteps[len(newSteps)-1])
	}

	return newSteps, nil
}

func CreateTimelapse(module Module, duration int) (timelapse [][]byte, err error) {
	listVcc, err := generateSteps(module.Scenario.Voltage, duration)
	if err != nil {
		return nil, err
	}

	listTemp, err := generateSteps(module.Scenario.Temperature, duration)
	if err != nil {
		return nil, err
	}

	listTxPower, err := generateSteps(module.Scenario.TxPower, duration)
	if err != nil {
		return nil, err
	}

	listRxPower, err := generateSteps(module.Scenario.RxPower, duration)
	if err != nil {
		return nil, err
	}

	listOsnr, err := generateSteps(module.Scenario.Osnr, duration)
	if err != nil {
		return nil, err
	}

	for i := 0; i < duration; i++ {
		step := make([]byte, 0)
		step = append(step, module.PageLow(listTemp[i], listVcc[i])...)
		step = append(step, module.Page00h()...)
		step = append(step, module.Page01h()...)
		step = append(step, module.Page02h()...)
		step = append(step, module.Page04h()...)
		step = append(step, module.Page11h(listTxPower[i], listRxPower[i])...)
		step = append(step, module.Page12h()...)
		step = append(step, module.Page25h(listOsnr[i], listTemp[i])...)

		timelapse = append(timelapse, step)
	}

	return
}

func SaveToFile(outputPath string, moduleName string, data [][]byte) error {
	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		return err
	}

	for timestamp, value := range data {
		filePath := path.Join(outputPath, fmt.Sprintf("%s-%09d", moduleName, timestamp))
		if _, err := os.Create(filePath); err != nil {
			return err
		}

		if err := os.WriteFile(filePath, value, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
