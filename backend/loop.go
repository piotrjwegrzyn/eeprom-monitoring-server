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

		SshSessions := []DeviceSignalsHolder{}

		for device := range devices {
			signals := DeviceSignalsHolder{&devices[device], make(chan bool), make(chan bool, 1)}
			SshSessions = append(SshSessions, signals)
			go MonitorData(&devices[device], &serverConfig.Influx, signals.SignalIn, signals.SignalOut, serverConfig.Intervals.SshQueryInt)
		}

		log.Printf("SQL interval sleep for %d seconds\n", serverConfig.Intervals.SqlQueryInt)
		time.Sleep(time.Duration(serverConfig.Intervals.SqlQueryInt) * time.Second)

		for device := range devices {
			select {
			case errorQuit := <-SshSessions[device].SignalOut:
				if errorQuit {
					devices[device].Status = 1
				}
			default:
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
