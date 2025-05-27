package monitor

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
	SleepTime      int `envconfig:"MONITOR_SLEEP_TIME_SECONDS" default:"30"`
	SSHTimeout     int `envconfig:"MONITOR_SSH_TIMEOUT_SECONDS" default:"10"`
	MaxConcurrency int `envconfig:"MONITOR_MAX_CONCURRENCY" default:"10"`
}

type Monitor struct {
	config Config
	db     *storage.DB
	influx *influx.Client
}

func New(cfg Config, db *storage.DB, influx *influx.Client) *Monitor {
	return &Monitor{config: cfg, db: db, influx: influx}
}

func (m *Monitor) Run(ctx context.Context) error {
	for {
		slog.InfoContext(ctx, fmt.Sprintf("waiting %d seconds", m.config.SleepTime))
		time.Sleep(time.Duration(m.config.SleepTime) * time.Second)

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		slog.InfoContext(ctx, "starting monitoring")

		devices, err := m.db.Devices(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error while getting devices", slog.Any("error", err))

			continue
		}

		streamDevices := make(chan storage.Device)

		wg := sync.WaitGroup{}
		wg.Add(m.config.MaxConcurrency)
		for range m.config.MaxConcurrency {
			go func() {
				for d := range streamDevices {
					remoteDev := newRemoteDevice(d, DefaultDecoder())

					status := m.monitorDevice(ctx, remoteDev)

					if err := m.updateStatus(ctx, &d, status); err != nil {
						slog.ErrorContext(ctx, "error while updating device", slog.Any("deviceID", d.ID), slog.Any("status", status))
					}
				}

				wg.Done()
			}()
		}

		for _, d := range devices {
			streamDevices <- d
		}
		close(streamDevices)

		wg.Wait()

		slog.InfoContext(ctx, "finished monitoring")
	}
}

func (m Monitor) monitorDevice(ctx context.Context, d remoteDevice) (status int8) {
	slog.InfoContext(ctx, "started device monitoring", slog.Any("deviceID", d.ID))

	auth, err := d.auth()
	if err != nil {
		slog.ErrorContext(ctx, "cannot parse key", slog.Any("deviceID", d.ID), slog.Any("error", err))

		return storage.StatusErrorKeyfile
	}

	client, err := d.sshClient(auth, m.config.SSHTimeout)
	if err != nil {
		slog.ErrorContext(ctx, "SSH client error", slog.Any("deviceID", d.ID), slog.Any("error", err))

		return storage.StatusErrorSSH
	}
	defer func() {
		if err := client.Close(); err != nil {
			slog.ErrorContext(ctx, "cannot close client connection", slog.Any("error", err))
		}
	}()

	slog.DebugContext(ctx, "created SSH client", slog.Any("deviceID", d.ID))

	interfaces, err := d.getInterfaces(client)
	if err != nil {
		slog.ErrorContext(ctx, "error with getting interfaces", slog.Any("deviceID", d.ID), slog.Any("error", err))

		return storage.StatusErrorSSH
	}

	slog.DebugContext(ctx, "detected interfaces", slog.Any("deviceID", d.ID), slog.Int("interfaces", len(interfaces)))

	failedRuns := 0
	for failedRuns < FailedRunsLimit {
		data, err := d.monitorInterfaces(client, interfaces)
		if err != nil {
			slog.WarnContext(ctx, "monitoring error", slog.Any("deviceID", d.ID), slog.Any("error", err))
			failedRuns += 1
			continue
		}

		for _, measurement := range data {
			m.influx.InsertMeasurements(d.Hostname, measurement.Interface, measurement.Measurement)
		}

		break
	}

	switch failedRuns {
	case 0:
		return storage.StatusOK
	case FailedRunsLimit:
		slog.WarnContext(ctx, "monitoring failed (error limit exceeded)", slog.Any("deviceID", d.ID))
		return storage.StatusErrorSSH
	default:
		return storage.StatusWarning
	}
}

func (m *Monitor) updateStatus(ctx context.Context, device *storage.Device, status int8) (err error) {
	device.LastStatus = status
	if device.LastStatus == storage.StatusOK {
		device.Connected = time.Now()
	}

	return m.db.UpdateDeviceStatus(ctx, *device)
}
