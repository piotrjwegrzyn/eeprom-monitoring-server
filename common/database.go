package common

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type Device struct {
	Id        int
	Hostname  string
	Ip        string
	Login     string
	Password  string
	Key       []byte
	Connected string
	Status    int
}

func (d *Device) GetStatusConnected() string {
	if d.Connected == "Never" && d.Status == 0 {
		return "STATUS UNDEFINED (NEVER CONNECTED)"
	} else if d.Status == 0 {
		return fmt.Sprintf("STATUS UNDEFINED (%s)", d.Connected)
	} else if d.Status == -1 {
		return "NO ROUTE TO HOST"
	} else if d.Status == 2 {
		return fmt.Sprintf("CREDENTIALS MISCONFIGURED (%s)", d.Connected)
	} else if d.Status == 3 {
		return "CREDENTIALS FAILED"
	} else {
		return fmt.Sprintf("STATUS OK (%s)", d.Connected)
	}
}

func InsertDevice(database *sql.DB, device Device) error {
	insert, err := database.Prepare("INSERT INTO devices(`hostname`, `ip`, `login`, `password`, `key`) VALUES(?,?,?,?,?)")
	if err != nil {
		return err
	}
	insert.Exec(device.Hostname, device.Ip, device.Login, device.Password, device.Key)
	log.Printf("Insert Device (%s) to database\n", device.Hostname)

	return nil
}

func UpdateDevice(database *sql.DB, device Device) error {
	update, err := database.Prepare("UPDATE devices SET `hostname`=?, `ip`=?, `login`=?, `password`=?, `key`=? WHERE `id`=?")
	if err != nil {
		return err
	}
	result, err := update.Exec(device.Hostname, device.Ip, device.Login, device.Password, device.Key, device.Id)
	if err != nil {
		return err
	}

	if affected, _ := result.RowsAffected(); affected == 1 {
		log.Printf("Updated device with ID: %d\n", device.Id)
	} else if affected == 0 {
		log.Printf("Nothing happen with update device with ID: %d (%d rows affected)\n", device.Id, affected)
	} else {
		return fmt.Errorf("%d rows affected during update device with ID: %d", affected, device.Id)
	}

	return nil
}

func DeleteDevice(database *sql.DB, id int) error {
	delete, err := database.Prepare("DELETE FROM devices WHERE `id`=?")
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

func GetDevice(database *sql.DB, id int) (Device, error) {
	devices, err := GetDevices(database)
	if err != nil {
		return Device{}, err
	}

	for d := range devices {
		if devices[d].Id == id {
			return devices[d], nil
		}
	}

	return Device{}, fmt.Errorf("no device with id: %d", id)
}

func GetDevices(database *sql.DB) (devices []Device, err error) {
	output, err := database.Query("SELECT * FROM devices;")
	if err != nil {
		return nil, err
	}

	for output.Next() {
		device := Device{}
		var pass sql.NullString
		if err := output.Scan(&device.Id, &device.Hostname, &device.Ip, &device.Login, &pass, &device.Key, &device.Connected, &device.Status); err != nil {
			log.Println(err)
		} else {
			device.Password = pass.String
			devices = append(devices, device)
		}
	}

	return
}

func ConnectToDatabase(config *DbConfig) (database *sql.DB, err error) {
	database, err = sql.Open("mysql", config.String())
	if err != nil {
		return
	}

	if err = database.Ping(); err != nil {
		return
	}

	log.Println("Connected to database")
	return
}
