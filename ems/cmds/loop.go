package cmds

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"pi-wegrzyn/ems/influx"
	"pi-wegrzyn/ems/storage"
)

type Config struct {
	StartupDelay   float32 `envconfig:"STARTUP_DELAY" default:"30.0"`
	SQLDelay       float32 `envconfig:"SQL_DELAY" default:"30.0"`
	SSHDelay       float32 `envconfig:"SQL_DELAY" default:"10.0"`
	MaxConcurrency int     `envconfig:"MAX_CONCURRENCY" default:"10"`
}

type server struct {
	config Config
	db     *storage.DB
	influx *influx.Client
}

func NewServer(cfg Config, db *storage.DB, influx *influx.Client) *server {
	return &server{config: cfg, db: db, influx: influx}
}

func (s *server) Loop(ctx context.Context) error {
	time.Sleep(time.Duration(s.config.StartupDelay) * time.Second)

	slog.InfoContext(ctx, "backend loop started")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		slog.InfoContext(ctx, fmt.Sprintf("waiting %.1f seconds", s.config.SQLDelay))
		timestamp := time.Now().Add(time.Second * time.Duration(s.config.SQLDelay))
		for time.Now().Before(timestamp) {
			continue
		}

		devices, err := s.db.Devices(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error while getting devices", slog.Any("error", err))

			continue
		}

		streamDevices := make(chan storage.Device)

		timeLimit := time.Now().Add(time.Second * time.Duration(s.config.SQLDelay-s.config.SSHDelay))
		timeSleep := time.Duration(s.config.SSHDelay) * time.Second

		wg := sync.WaitGroup{}
		wg.Add(s.config.MaxConcurrency)
		for range s.config.MaxConcurrency {
			go func() {
				for d := range streamDevices {
					remoteDev := NewRemoteDevice(d, defaultDecoder)

					status := Monitor(ctx, s.influx, timeLimit, timeSleep, remoteDev)

					if err := s.updateStatus(ctx, &d, status); err != nil {
						slog.ErrorContext(ctx, "error while updating device", slog.Any("deviceID", d.ID), slog.Any("status", status))
					}
				}

				wg.Done()
			}()
		}

		for _, d := range devices {
			streamDevices <- d
		}

		wg.Wait()
	}
}

func (s *server) updateStatus(ctx context.Context, device *storage.Device, status int8) (err error) {
	device.LastStatus = status
	if device.LastStatus == storage.STATUS_OK {
		device.Connected = time.Now()
	}

	return s.db.UpdateDeviceStatus(ctx, *device)
}
