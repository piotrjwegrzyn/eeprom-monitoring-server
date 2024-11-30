package cmds

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"pi-wegrzyn/storage"
	"pi-wegrzyn/utils"

	"golang.org/x/crypto/ssh"
)

const (
	CMD_SHOW_EEPROM           string = "show-eeprom %s"
	CMD_SHOW_FIBER_INTERFACES string = "show-fiber-interfaces"
	FAILED_RUNS_LIMIT         int    = 5
)

type remoteDevice struct {
	storage.Device
	decode     func([]byte) ([]byte, error)
	ExitSignal chan int8
}

func NewRemoteDevice(dev storage.Device, decode func([]byte) ([]byte, error)) *remoteDevice {
	return &remoteDevice{
		Device:     dev,
		decode:     decode,
		ExitSignal: make(chan int8, 1),
	}
}

func (d *remoteDevice) Monitor(ctx context.Context, influx *utils.Influx, timeLimit time.Time, timeSleep time.Duration) {
	slog.InfoContext(ctx, "started goroutine", slog.Any("deviceID", d.ID))

CLIENT_CREATION:
	auth, err := d.auth()
	if err != nil {
		slog.ErrorContext(ctx, "cannot parse key", slog.Any("deviceID", d.ID), slog.Any("error", err))
		d.ExitSignal <- storage.STATUS_ERROR_KEYFILE
		return
	}

	client, err := d.sshClient(auth)
	if err != nil {
		slog.ErrorContext(ctx, "SSH client error", slog.Any("deviceID", d.ID), slog.Any("error", err))
		d.ExitSignal <- storage.STATUS_ERROR_SSH
		return
	}
	defer client.Close()
	slog.InfoContext(ctx, "created SSH client", slog.Any("deviceID", d.ID))

	interfaces, err := d.getInterfaces(client)
	if err != nil {
		fmt.Printf("Error with getting interfaces: %v (device ID: %d)\n", err, d.ID)
		slog.ErrorContext(ctx, "error with getting interfaces", slog.Any("deviceID", d.ID), slog.Any("error", err))
		d.ExitSignal <- storage.STATUS_ERROR_SSH
		return
	}
	slog.InfoContext(ctx, fmt.Sprintf("detected %d interface(s)", len(interfaces)), slog.Any("deviceID", d.ID))

	failedRuns := 0
	for time.Now().Before(timeLimit) {
		err := d.monitorInterfaces(client, interfaces, influx)
		if err != nil {
			slog.WarnContext(ctx, "monitoring error", slog.Any("deviceID", d.ID), slog.Any("error", err))
			failedRuns += 1
		}

		if failedRuns > FAILED_RUNS_LIMIT {
			slog.WarnContext(ctx, fmt.Sprintf("Error limit exceeded (%d errors), trying to reconnect", failedRuns), slog.Any("deviceID", d.ID))
			if err := client.Close(); err != nil {
				slog.ErrorContext(ctx, "closing session error", slog.Any("deviceID", d.ID), slog.Any("error", err))
				d.ExitSignal <- storage.STATUS_ERROR_SSH
				return
			}

			goto CLIENT_CREATION
		}

		time.Sleep(timeSleep)
	}

	if failedRuns != 0 {
		d.ExitSignal <- storage.STATUS_WARNING
	}

	d.ExitSignal <- storage.STATUS_OK
}

func (d *remoteDevice) auth() ([]ssh.AuthMethod, error) {
	if len(d.Keyfile) == 0 {
		return []ssh.AuthMethod{ssh.Password(d.Password)}, nil
	}

	signer, err := ssh.ParsePrivateKey(d.Keyfile)
	if err != nil {
		return nil, err
	}

	return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
}

func (d *remoteDevice) sshClient(auth []ssh.AuthMethod) (*ssh.Client, error) {
	sshCfg := &ssh.ClientConfig{
		Auth:            auth,
		User:            d.Login,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", d.IPAddress+":22", sshCfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (d *remoteDevice) getInterfaces(client *ssh.Client) ([]string, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	got, err := session.CombinedOutput(CMD_SHOW_FIBER_INTERFACES)
	if err != nil {
		return nil, err
	}

	infs := strings.SplitN(string(got), "\n", -1)

	return infs[:len(infs)-1], nil
}

func (d *remoteDevice) monitorInterfaces(client *ssh.Client, interfaces []string, influx *utils.Influx) (err error) {
	for _, inf := range interfaces {
		session, err2 := client.NewSession()
		if err2 != nil {
			err = errors.Join(err, fmt.Errorf("%v (interface: %s)", err2, inf))
			continue
		}
		defer session.Close()

		got, err2 := session.CombinedOutput(fmt.Sprintf(CMD_SHOW_EEPROM, inf))
		if err2 != nil {
			err = errors.Join(err, fmt.Errorf("%v (interface: %s)", err2, inf))
			continue
		}

		interfaceData, err2 := d.processData(got)
		if err2 != nil {
			err = errors.Join(err, fmt.Errorf("%v (interface: %s)", err2, inf))
			continue
		}

		influx.Insert(d.Hostname, inf, &interfaceData)
	}

	return
}

func (d *remoteDevice) processData(input []byte) (utils.InterfaceData, error) {
	decoded, err := d.decode(input)
	if err != nil {
		return utils.InterfaceData{}, err
	}

	return utils.InterfaceData{
		Temperature: temperature(decoded),
		Voltage:     voltage(decoded),
		TxPower:     txPower(decoded),
		RxPower:     rxPower(decoded),
		Osnr:        osnr(decoded),
	}, nil
}
