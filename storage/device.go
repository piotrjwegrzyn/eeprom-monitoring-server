package storage

import (
	"fmt"
	"time"
)

const (
	STATUS_UNDEFINED = iota - 1
	STATUS_OK
	STATUS_ERROR_SSH
	STATUS_ERROR_KEYFILE
	STATUS_WARNING
)

type Device struct {
	ID         uint32
	Hostname   string
	IPAddress  string
	Login      string
	Password   string
	Keyfile    []byte
	Connected  time.Time
	LastStatus int8
}

func (d *Device) StatusConnected() string {
	connected := d.Connected.Format(time.RFC3339)

	switch d.LastStatus {
	case STATUS_UNDEFINED:
		return "STATUS UNDEFINED (NEVER CONNECTED)"
	case STATUS_OK:
		return fmt.Sprintf("STATUS OK (last connection: %s)", connected)
	case STATUS_ERROR_SSH:
		return fmt.Sprintf("SSH SESSION ERROR (last connection: %s)", connected)
	case STATUS_ERROR_KEYFILE:
		return fmt.Sprintf("KEYFILE ERROR (last connection: %s)", connected)
	case STATUS_WARNING:
		return fmt.Sprintf("SOME ERRORS OCCURRED (last connection: %s)", connected)
	default:
		return "STATUS UNKNOWN"
	}
}
