package cmds

import (
	"errors"
	"fmt"
	"log"
	"time"

	"pi-wegrzyn/utils"
)

type server struct {
	config   utils.Config
	database utils.Database
	remotes  []*remoteDevice
}

func NewServer(cfg *utils.Config) *server {
	return &server{config: *cfg}
}

func (s *server) Loop() error {
	log.Printf("Startup delay set for %.1f seconds\n", s.config.Delays.Startup)
	time.Sleep(time.Duration(s.config.Delays.Startup) * time.Second)

	log.Println("Backend module started")
	for {
		var err error
		s.database, err = utils.ConnectToDatabase(&s.config.MySQL)
		if err != nil {
			return err
		}
		defer s.database.Close()

		if err = s.prepareRemotes(); err != nil {
			return err
		}

		log.Printf("SQL delay set to %.1f seconds\n", s.config.Delays.SQL)
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

func (s *server) prepareRemotes() error {
	devices, err := s.database.GetDevices()
	if err != nil {
		return err
	}

	s.remotes = []*remoteDevice{}
	for _, dev := range devices {
		c := NewRemoteDevice(dev, defaultDecoder)

		s.remotes = append(s.remotes, c)

		timeLimit := time.Now().Add(time.Second * time.Duration(s.config.Delays.SQL-s.config.Delays.SSH))
		timeSleep := time.Duration(s.config.Delays.SSH) * time.Second

		go c.Monitor(&s.config.Influx, timeLimit, timeSleep)
	}

	log.Printf("Prepared %d remote handlers\n", len(s.remotes))
	return nil
}

func (s *server) checkSignalsLoop() (err error) {
	for _, r := range s.remotes {
		select {
		case got := <-r.ExitSignal:
			log.Printf("goroutine ended with status: %d (device ID: %d)\n", got, r.ID)
			r.Status = got
			if got == utils.STATUS_OK {
				r.Connected = time.Now().Format("2006-01-02 15:04:05")
			}
		default:
			log.Printf("No exit of goroutine (device ID: %d)\n", r.ID)
			r.Status = utils.STATUS_OK
			r.Connected = time.Now().Format("2006-01-02 15:04:05")
		}

		if err2 := s.database.UpdateDeviceStatus(*r); err2 != nil {
			errors.Join(err, fmt.Errorf("error while updating device: %v (device ID: %d)", err2, r.ID))
		}
	}

	return
}
