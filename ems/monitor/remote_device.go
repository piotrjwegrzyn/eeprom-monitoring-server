package monitor

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"pi-wegrzyn/ems/influx"
	"pi-wegrzyn/ems/storage"

	"golang.org/x/crypto/ssh"
)

const (
	CmdShowEEPROM          string = "show-eeprom %s"
	CmdShowFiberInterfaces string = "show-fiber-interfaces"
	FailedRunsLimit        int    = 5
)

type interfaceMeasurement struct {
	influx.Measurement

	Interface string
}

type remoteDevice struct {
	storage.Device
	decodeFunc Decoder
}

func newRemoteDevice(dev storage.Device, decoder func([]byte) (Eeprom, error)) remoteDevice {
	return remoteDevice{
		Device:     dev,
		decodeFunc: decoder,
	}
}

func (d remoteDevice) auth() ([]ssh.AuthMethod, error) {
	if len(d.Keyfile) == 0 {
		return []ssh.AuthMethod{ssh.Password(d.Password)}, nil
	}

	signer, err := ssh.ParsePrivateKey(d.Keyfile)
	if err != nil {
		return nil, err
	}

	return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
}

func (d remoteDevice) sshClient(auth []ssh.AuthMethod, timeout int) (*ssh.Client, error) {
	sshCfg := &ssh.ClientConfig{
		Auth:            auth,
		User:            d.Login,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	}

	return ssh.Dial("tcp", d.IPAddress+":22", sshCfg)
}

func (d remoteDevice) getInterfaces(client *ssh.Client) ([]string, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := session.Close(); err != nil {
			slog.Error("cannot close session", slog.Any("error", err))
		}
	}()

	got, err := session.CombinedOutput(CmdShowFiberInterfaces)
	if err != nil {
		return nil, err
	}

	infs := strings.Split(string(got), "\n")

	return infs[:len(infs)-1], nil
}

func (d remoteDevice) monitorInterfaces(client *ssh.Client, interfaces []string) (measurements []interfaceMeasurement, err error) {
	for _, inf := range interfaces {
		session, err2 := client.NewSession()
		if err2 != nil {
			err = errors.Join(err, fmt.Errorf("%v (interface: %s)", err2, inf))
			continue
		}
		defer func() {
			if err := session.Close(); err != nil {
				slog.Error("cannot close session", slog.Any("error", err))
			}
		}()

		got, err2 := session.CombinedOutput(fmt.Sprintf(CmdShowEEPROM, inf))
		if err2 != nil {
			err = errors.Join(err, fmt.Errorf("%v (interface: %s)", err2, inf))
			continue
		}

		ifData, err2 := d.processData(got)
		if err2 != nil {
			err = errors.Join(err, fmt.Errorf("%v (interface: %s)", err2, inf))
			continue
		}

		measurements = append(measurements, interfaceMeasurement{
			Measurement: ifData,
			Interface:   inf,
		})
	}

	return measurements, err
}

func (d remoteDevice) processData(input []byte) (influx.Measurement, error) {
	decoded, err := d.decodeFunc(input)
	if err != nil {
		return influx.Measurement{}, err
	}

	return influx.Measurement{
		Temperature: decoded.Temperature(),
		Voltage:     decoded.Voltage(),
		TxPower:     decoded.TxPower(),
		RxPower:     decoded.RxPower(),
		OSNR:        decoded.Osnr(),
	}, nil
}
