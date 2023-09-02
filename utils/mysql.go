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

type Database struct {
	*sql.DB
}

type dbDevice interface {
	GetDevice() Device
}

func (db *Database) InsertDevice(dev dbDevice) error {
	d := dev.GetDevice()

	insert, err := db.Prepare("INSERT INTO devices(`hostname`, `ip`, `login`, `password`, `key`) VALUES(?,?,?,?,?)")
	if err != nil {
		return err
	}

	if _, err := insert.Exec(d.Hostname, d.IP, d.Login, d.Password, d.Key); err != nil {
		return err
	}

	log.Printf("Inserted Device (%s) to database\n", d.Hostname)
	return nil
}

func (db *Database) UpdateDevice(dev dbDevice) error {
	d := dev.GetDevice()

	update, err := db.Prepare("UPDATE devices SET `hostname`=?, `ip`=?, `login`=?, `password`=?, `key`=?, `status`=?, `connected`=? WHERE `id`=?")
	if err != nil {
		return err
	}
	result, err := update.Exec(d.Hostname, d.IP, d.Login, d.Password, d.Key, d.Status, d.Connected, d.ID)
	if err != nil {
		return err
	}

	if affected, _ := result.RowsAffected(); affected == 1 {
		log.Printf("Updated (device ID: %d)\n", d.ID)
	} else if affected == 0 {
		log.Printf("Nothing happen with update (device ID: %d) (%d rows affected)\n", d.ID, affected)
	} else {
		return fmt.Errorf("%d rows affected during update (device ID: %d)", affected, d.ID)
	}

	return nil
}

func (db *Database) UpdateDeviceStatus(dev dbDevice) error {
	d := dev.GetDevice()

	current, err := db.GetDevice(d.ID)
	if err != nil {
		return err
	}
	current.Status, current.Connected = d.Status, d.Connected

	err = db.UpdateDevice(current)
	if err != nil {
		return err
	}

	log.Printf("Updated (device ID: %d)\n", current.ID)
	return nil
}

func (db *Database) DeleteDevice(id int) error {
	delete, err := db.Prepare("DELETE FROM devices WHERE `id`=?")
	if err != nil {
		return err
	}

	result, err := delete.Exec(id)
	if err != nil {
		return err
	}

	if affected, _ := result.RowsAffected(); affected == 1 {
		log.Printf("Deleted (device ID: %d)\n", id)
	} else if affected == 0 {
		log.Printf("Nothing happen with deletion (device ID: %d) (%d rows affected)\n", id, affected)
	} else {
		return fmt.Errorf("%d rows affected during deletion (device ID: %d)", affected, id)
	}

	return nil
}

func (db *Database) GetDevice(id int) (Device, error) {
	devices, err := db.GetDevices()
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

func (db *Database) GetDevices() (devices []Device, err error) {
	output, err := db.Query("SELECT * FROM devices;")
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

	log.Printf("Got %d devices from database\n", len(devices))
	return
}

func ConnectToDatabase(config *MySQL) (db Database, err error) {
	opened, err := sql.Open("mysql", config.String())
	if err != nil {
		return
	}

	if err = opened.Ping(); err != nil {
		return
	}

	log.Println("Connected to database")
	return Database{opened}, nil
}
