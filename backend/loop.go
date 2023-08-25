package main

import (
	"log"
	"time"

	"pi-wegrzyn/utils"
)

type DeviceSignalsHolder struct {
	DevicePointer *utils.Device
	SignalIn      chan bool
	SignalOut     chan bool
}

func StartLoop(serverConfig *utils.Config) error {
	for {
		database, err := utils.ConnectToDatabase(&serverConfig.Database)
		if err != nil {
			return err
		}
		defer database.Close()

		devices, err := utils.GetDevices(database)
		if err != nil {
			return err
		}
		log.Printf("Got %d devices from database\n", len(devices))

		SshSessions := []DeviceSignalsHolder{}

		for device := range devices {
			signals := DeviceSignalsHolder{&devices[device], make(chan bool, 1), make(chan bool, 1)}
			SshSessions = append(SshSessions, signals)
			go MonitorData(&devices[device], &serverConfig.Influx, signals.SignalIn, signals.SignalOut, serverConfig.Intervals.SshQueryInt)
		}

		log.Printf("SQL check interval set to %d seconds\n", serverConfig.Intervals.SqlQueryInt)

		timestamp := time.Now().Add(time.Second * time.Duration(serverConfig.Intervals.SqlQueryInt))
		for time.Now().Before(timestamp) {
			continue
		}

		for device := range devices {
			select {
			case <-SshSessions[device].SignalOut:
				log.Printf("Detected error signal from device with ID: %d\n", devices[device].Id)
				if devices[device].Status != 2 {
					devices[device].Status = 1
				}
			default:
				log.Printf("Sending stop signal to device with ID: %d\n", devices[device].Id)
				SshSessions[device].SignalIn <- true
				if devices[device].Status != 1 && devices[device].Status != 2 {
					devices[device].Status = 0
				}
				devices[device].Connected = time.Now().Format("2006-01-02 15:04:05")
			}

			err = utils.UpdateDeviceStatus(database, devices[device])
			if err != nil {
				log.Printf("Error while updating device with ID: %d (%s)\n", devices[device].Id, err)
			}
		}
	}
}
