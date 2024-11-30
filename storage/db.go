//go:generate sqlc generate -f sqlc/config.yaml

package storage

import (
	"context"
	"database/sql"
	"time"

	sqlc "pi-wegrzyn/storage/sqlc/generated"
)

type DB struct {
	q *sqlc.Queries
}

func New(dbConn *sql.DB) *DB {
	return &DB{q: sqlc.New(dbConn)}
}

func (d *DB) CreateDevice(ctx context.Context, device Device) error {
	createParams := sqlc.CreateDeviceParams{
		Hostname: device.Hostname,
		Ip:       device.IPAddress,
		Login:    device.Login,
		Passwd: sql.NullString{
			String: device.Password,
		},
		Keyfile: sql.NullString{
			String: string(device.Keyfile),
		},
		Connected: time.Now(),
	}

	return d.q.CreateDevice(ctx, createParams)
}

func (d *DB) Device(ctx context.Context, id uint) (Device, error) {
	dbDevice, err := d.q.Device(ctx, uint32(id))
	if err != nil {
		return Device{}, err
	}

	return Device{
		ID:         dbDevice.ID,
		Hostname:   dbDevice.Hostname,
		IPAddress:  dbDevice.Ip,
		Login:      dbDevice.Login,
		Password:   dbDevice.Passwd.String,
		Keyfile:    []byte(dbDevice.Keyfile.String),
		Connected:  dbDevice.Connected,
		LastStatus: int8(dbDevice.LastStatus),
	}, nil
}

func (d *DB) Devices(ctx context.Context) ([]Device, error) {
	dbDevices, err := d.q.Devices(ctx)
	if err != nil {
		return nil, err
	}

	devices := make([]Device, 0, len(dbDevices))
	for _, dev := range dbDevices {
		devices = append(devices, Device{
			ID:         dev.ID,
			Hostname:   dev.Hostname,
			Login:      dev.Login,
			IPAddress:  dev.Ip,
			Password:   dev.Passwd.String,
			Keyfile:    []byte(dev.Keyfile.String),
			Connected:  dev.Connected,
			LastStatus: int8(dev.LastStatus),
		})
	}

	return devices, nil
}

func (d *DB) UpdateDevice(ctx context.Context, device Device) error {
	updateParams := sqlc.UpdateDeviceParams{
		ID:       device.ID,
		Hostname: device.Hostname,
		Ip:       device.IPAddress,
		Login:    device.Login,
		Passwd: sql.NullString{
			String: device.Password,
		},
		Keyfile: sql.NullString{
			String: string(device.Keyfile),
		},
		Connected:  device.Connected,
		LastStatus: int32(device.LastStatus),
	}
	return d.q.UpdateDevice(ctx, updateParams)
}

func (d *DB) UpdateDeviceStatus(ctx context.Context, device Device) error {
	updateParams := sqlc.UpdateDeviceStatusParams{
		ID:         device.ID,
		Connected:  device.Connected,
		LastStatus: int32(device.LastStatus),
	}
	return d.q.UpdateDeviceStatus(ctx, updateParams)
}

func (d *DB) DeleteDevice(ctx context.Context, id uint) error {
	return d.q.DeleteDevice(ctx, uint32(id))
}
