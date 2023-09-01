package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type MySQL struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Protocol string `yaml:"protocol"`
	Address  string `yaml:"address"`
	DBName   string `yaml:"dbname"`
}

func (m *MySQL) String() string {
	return fmt.Sprintf("%s:%s@%s(%s)/%s", m.Username, m.Password, m.Protocol, m.Address, m.DBName)
}

type database struct {
	*sql.DB
}

func (d *database) InsertDevice(device Device) error {
	insert, err := d.Prepare("INSERT INTO devices(`hostname`, `ip`, `login`, `password`, `key`) VALUES(?,?,?,?,?)")
	if err != nil {
		return err
	}
	insert.Exec(device.Hostname, device.IP, device.Login, device.Password, device.Key)
	log.Printf("Insert Device (%s) to database\n", device.Hostname)

	return nil
}

func (d *database) UpdateDevice(device Device) error {
	update, err := d.Prepare("UPDATE devices SET `hostname`=?, `ip`=?, `login`=?, `password`=?, `key`=?, `status`=?, `connected`=? WHERE `id`=?")
	if err != nil {
		return err
	}
	result, err := update.Exec(device.Hostname, device.IP, device.Login, device.Password, device.Key, device.Status, device.Connected, device.ID)
	if err != nil {
		return err
	}

	if affected, _ := result.RowsAffected(); affected == 1 {
		log.Printf("Updated device with ID: %d\n", device.ID)
	} else if affected == 0 {
		log.Printf("Nothing happen with update device with ID: %d (%d rows affected)\n", device.ID, affected)
	} else {
		return fmt.Errorf("%d rows affected during update device with ID: %d", affected, device.ID)
	}

	return nil
}

func (d *database) UpdateDeviceStatus(device Device) error {
	currentDevice, err := d.GetDevice(device.ID)
	if err != nil {
		return err
	}
	currentDevice.Status, currentDevice.Connected = device.Status, device.Connected

	err = d.UpdateDevice(currentDevice)
	if err != nil {
		return err
	}

	return nil
}

func (d *database) DeleteDevice(id int) error {
	delete, err := d.Prepare("DELETE FROM devices WHERE `id`=?")
	if err != nil {
		return err
	}

	result, err := delete.Exec(id)
	if err != nil {
		return err
	}

	if affected, _ := result.RowsAffected(); affected == 1 {
		log.Printf("Deleted device with ID: %d\n", id)
	} else if affected == 0 {
		log.Printf("Nothing happen with deletion device with ID: %d (%d rows affected)\n", id, affected)
	} else {
		return fmt.Errorf("%d rows affected during deletion device with ID: %d", affected, id)
	}

	return nil
}

func (d *database) GetDevice(id int) (Device, error) {
	devices, err := d.GetDevices()
	if err != nil {
		return Device{}, err
	}

	for dev := range devices {
		if devices[dev].ID == id {
			return devices[dev], nil
		}
	}

	return Device{}, fmt.Errorf("no device with id: %d", id)
}

func (d *database) GetDevices() (devices []Device, err error) {
	output, err := d.Query("SELECT * FROM devices;")
	if err != nil {
		return nil, err
	}

	for output.Next() {
		device := Device{}
		var pass sql.NullString
		if err2 := output.Scan(&device.ID, &device.Hostname, &device.IP, &device.Login, &pass, &device.Key, &device.Connected, &device.Status); err2 != nil {
			err = errors.Join(err, err2)
			continue
		}

		device.Password = pass.String
		devices = append(devices, device)
	}

	return
}

func ConnectToDatabase(config *MySQL) (db database, err error) {
	opened, err := sql.Open("mysql", config.String())
	if err != nil {
		return
	}

	if err = opened.Ping(); err != nil {
		return
	}

	log.Println("Connected to database")
	return database{opened}, nil
}
