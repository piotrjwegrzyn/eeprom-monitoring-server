package utils

import "fmt"

const (
	STATUS_UNDEFINED = iota - 1
	STATUS_OK
	STATUS_ERROR_SSH
	STATUS_ERROR_KEYFILE
	STATUS_UNKNOWN
)

type Device struct {
	ID        int
	Hostname  string
	IP        string
	Login     string
	Password  string
	Key       []byte
	Connected string
	Status    int8
}

func (d Device) GetDevice() Device {
	return d
}

func (d *Device) StatusConnected() string {
	switch d.Status {
	case STATUS_UNDEFINED:
		return "STATUS UNDEFINED (NEVER CONNECTED)"
	case STATUS_OK:
		return fmt.Sprintf("STATUS OK (last connection: %s)", d.Connected)
	case STATUS_ERROR_SSH:
		return fmt.Sprintf("SSH SESSION ERROR (last connection: %s)", d.Connected)
	case STATUS_ERROR_KEYFILE:
		return fmt.Sprintf("KEYFILE ERROR (last connection: %s)", d.Connected)
	default:
		return fmt.Sprintf("STATUS UNKNOWN (last connection: %s)", d.Connected)
	}
}
