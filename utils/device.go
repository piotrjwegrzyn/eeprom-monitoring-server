package utils

import "fmt"

type Device struct {
	ID        int
	Hostname  string
	IP        string
	Login     string
	Password  string
	Key       []byte
	Connected string
	Status    int
}

func (d *Device) StatusConnected() string {
	switch d.Status {
	case -1:
		return "STATUS UNDEFINED (NEVER CONNECTED)"
	case 0:
		return fmt.Sprintf("STATUS OK (last connection: %s)", d.Connected)
	case 1:
		return fmt.Sprintf("SSH SESSION ERROR (last connection: %s)", d.Connected)
	case 2:
		return fmt.Sprintf("CREDENTIALS MISCONFIGURED (last connection: %s)", d.Connected)
	default:
		return fmt.Sprintf("STATUS UNKNOWN (last connection: %s)", d.Connected)
	}
}
