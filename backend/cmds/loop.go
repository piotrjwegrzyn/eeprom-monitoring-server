package cmds

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"pi-wegrzyn/backend/influx"
	"pi-wegrzyn/storage"
)

type Delays struct {
	Startup float32 `yaml:"startup"`
	SQL     float32 `yaml:"sql"`
	SSH     float32 `yaml:"ssh"`
}

type Config struct {
	Delays Delays `yaml:"delays"`
}

type server struct {
	config  Config
	db      *storage.DB
	influx  *influx.Client
	remotes []*remoteDevice
}

func NewServer(cfg Config, db *storage.DB, influx *influx.Client) *server {
	return &server{config: cfg, db: db, influx: influx}
}

func (s *server) Loop(ctx context.Context) error {
	slog.InfoContext(ctx, fmt.Sprintf("startup delay set for %.1f seconds", s.config.Delays.Startup))
	time.Sleep(time.Duration(s.config.Delays.Startup) * time.Second)

	slog.InfoContext(ctx, "backend module started")
	for {
		if err := s.prepareRemotes(ctx); err != nil {
			return err
		}

		slog.InfoContext(ctx, fmt.Sprintf("SQL delay set to %.1f seconds", s.config.Delays.SQL))
		timestamp := time.Now().Add(time.Second * time.Duration(s.config.Delays.SQL))
		for time.Now().Before(timestamp) {
			continue
		}

		if err := s.checkSignalsLoop(ctx); err != nil {
			slog.WarnContext(ctx, "error occurred", slog.Any("error", err))
		}
	}
}

func (s *server) prepareRemotes(ctx context.Context) error {
	devices, err := s.db.Devices(ctx)
	if err != nil {
		return err
	}

	s.remotes = []*remoteDevice{}
	for _, dev := range devices {
		c := NewRemoteDevice(dev, defaultDecoder)

		s.remotes = append(s.remotes, c)

		timeLimit := time.Now().Add(time.Second * time.Duration(s.config.Delays.SQL-s.config.Delays.SSH))
		timeSleep := time.Duration(s.config.Delays.SSH) * time.Second

		go c.Monitor(ctx, s.influx, timeLimit, timeSleep)
	}

	slog.InfoContext(ctx, fmt.Sprintf("prepared %d remote handlers", len(s.remotes)))
	return nil
}

func (s *server) checkSignalsLoop(ctx context.Context) (err error) {
	for _, r := range s.remotes {
		select {
		case got := <-r.ExitSignal:
			slog.InfoContext(ctx, "goroutine ended", slog.Any("deviceID", r.ID), slog.Any("status", got))
			r.LastStatus = got
			if got == storage.STATUS_OK {
				r.Connected = time.Now()
			}
		default:
			slog.InfoContext(ctx, "no exit of goroutine", slog.Any("deviceID", r.ID))
			r.LastStatus = storage.STATUS_OK
			r.Connected = time.Now()
		}

		if err2 := s.db.UpdateDeviceStatus(ctx, r.Device); err2 != nil {
			err = errors.Join(err, fmt.Errorf("error while updating device: %v (device ID: %d)", err2, r.ID))
		}
	}

	return err
}
