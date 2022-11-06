package main

import (
	"log"
	"time"

	common "pi-wegrzyn/common"
)

type DeviceSignalsHolder struct {
	DevicePointer *common.Device
	SignalIn      chan bool
	SignalOut     chan bool
}

func StartLoop(serverConfig *common.Config) error {
	for {
		database, err := common.ConnectToDatabase(&serverConfig.Database)
		if err != nil {
			return err
		}
		defer database.Close()

		devices, err := common.GetDevices(database)
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

		log.Printf("SQL interval sleep for %d seconds\n", serverConfig.Intervals.SqlQueryInt)

		for i := 0; i < serverConfig.Intervals.SqlQueryInt; i++ {
			time.Sleep(time.Second)
		}

		for device := range devices {
			select {
			case errorQuit := <-SshSessions[device].SignalOut:
				log.Printf("Detected error signal on device with ID: %d\n", devices[device].Id)
				if errorQuit {
					devices[device].Status = 1
				}
			default:
				log.Printf("Sending stop signal to device with ID: %d\n", devices[device].Id)
				SshSessions[device].SignalIn <- true
				if devices[device].Status != 2 {
					devices[device].Status = 10
				}
				devices[device].Connected = time.Now().Format("2006-01-02 15:04:05")
			}

			common.UpdateDevice(database, devices[device])
		}
	}
}
