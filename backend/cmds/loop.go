package cmds

import (
	"errors"
	"fmt"
	"log"
	"time"

	"pi-wegrzyn/utils"
)

type server struct {
	config     utils.Config
	database   utils.Database
	connectors []*connector
}

func NewServer(cfg *utils.Config) *server {
	return &server{config: *cfg}
}

type connector struct {
	remoteDevice
	In  chan bool
	Out chan bool
}

func (s *server) prepareSignalHolders() error {
	devices, err := s.database.GetDevices()
	if err != nil {
		return err
	}

	for _, dev := range devices {
		c := connector{
			remoteDevice{dev},
			make(chan bool, 1),
			make(chan bool, 1),
		}

		s.connectors = append(s.connectors, &c)

		go c.MonitorData(&s.config.Influx, c.In, c.Out, s.config.Delays.SSH)
	}

	log.Printf("Prepared %d signallers\n", len(s.connectors))
	return nil
}

func (s *server) Loop() error {
	log.Printf("Startup delay set for %d seconds\n", s.config.Delays.Startup)
	time.Sleep(time.Duration(s.config.Delays.Startup) * time.Second)

	log.Println("Backend module started")
	for {
		var err error
		s.database, err = utils.ConnectToDatabase(&s.config.MySQL)
		if err != nil {
			return err
		}
		defer s.database.Close()

		if err = s.prepareSignalHolders(); err != nil {
			return err
		}

		log.Printf("SQL delay set to %d seconds\n", s.config.Delays.SQL)
		timestamp := time.Now().Add(time.Second * time.Duration(s.config.Delays.SQL))
		for time.Now().Before(timestamp) {
			continue
		}

		err = s.checkSignalsLoop()
		if err != nil {
			log.Printf("Error(s) occurred: %v\n", err)
		}
	}
}

func (s *server) checkSignalsLoop() (err error) {
	for _, c := range s.connectors {
		select {
		case <-c.Out:
			log.Printf("Detected error signal from device with ID: %d\n", c.ID)
			if c.Status != utils.STATUS_ERROR_MISCONFIGURATION {
				c.Status = utils.STATUS_ERROR_SSH
			}
		default:
			log.Printf("Sending stop signal to device with ID: %d\n", c.ID)
			c.In <- true

			if c.Status != utils.STATUS_ERROR_SSH && c.Status != utils.STATUS_ERROR_MISCONFIGURATION {
				c.Status = utils.STATUS_OK
			}

			c.Connected = time.Now().Format("2006-01-02 15:04:05")
		}

		err2 := s.database.UpdateDeviceStatus(*c)
		if err2 != nil {
			errors.Join(err, fmt.Errorf("error while updating device with ID: %d (%v)", c.ID, err2))
		}
	}

	return
}
