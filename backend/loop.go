package main

import (
	common "pi-wegrzyn/common"
	"time"
)

var SshSessions map[*common.Device]chan int

func StartLoop(serverConfig *common.Config) error {

	database, err := common.ConnectToDatabase(&serverConfig.Database)
	if err != nil {
		return err
	}

	devices, err := common.GetDevices(database)
	if err != nil {
		return err
	}
	SshSessions = make(map[*common.Device]chan int)

	for device := range devices {
		SshSessions[&devices[device]] = make(chan int)
		go MonitorData(&devices[device], SshSessions[&devices[device]], serverConfig.Intervals.SshQueryInt)
	}

	time.Sleep(10 * time.Second)

	for device := range devices {
		SshSessions[&devices[device]] <- 1
	}

	return nil
}
