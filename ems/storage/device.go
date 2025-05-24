package storage

import (
	"fmt"
	"net"
	"time"
)

const (
	StatusUndefined = iota - 1
	StatusOK
	StatusErrorSSH
	StatusErrorKeyfile
	StatusWarning
)

type Device struct {
	ID         uint
	Hostname   string
	IPAddress  string
	Login      string
	Password   string
	Keyfile    []byte
	Connected  time.Time
	LastStatus int8
}

func (d *Device) IPVersion() int {
	ip := net.ParseIP(d.IPAddress)
	if ip.To4() != nil {
		return 4
	}

	return 6
}

func (d *Device) StatusConnected() string {
	connected := d.Connected.Format(time.RFC3339)

	switch d.LastStatus {
	case StatusUndefined:
		return "STATUS UNDEFINED (NEVER CONNECTED)"
	case StatusOK:
		return fmt.Sprintf("STATUS OK (last connection: %s)", connected)
	case StatusErrorSSH:
		return fmt.Sprintf("SSH SESSION ERROR (last connection: %s)", connected)
	case StatusErrorKeyfile:
		return fmt.Sprintf("KEYFILE ERROR (last connection: %s)", connected)
	case StatusWarning:
		return fmt.Sprintf("SOME ERRORS OCCURRED (last connection: %s)", connected)
	default:
		return "STATUS UNKNOWN"
	}
}
