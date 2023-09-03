package cmds

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"pi-wegrzyn/utils"

	"golang.org/x/crypto/ssh"
)

const (
	CMD_SHOW_EEPROM           string = "show-eeprom %s"
	CMD_SHOW_FIBER_INTERFACES string = "show-fiber-interfaces"
	FAILED_RUNS_LIMIT         int    = 5
)

type remoteDevice struct {
	utils.Device
	decode     func([]byte) ([]byte, error)
	ExitSignal chan int8
}

func NewRemoteDevice(dev utils.Device, decode func([]byte) ([]byte, error)) *remoteDevice {
	return &remoteDevice{
		Device:     dev,
		decode:     decode,
		ExitSignal: make(chan int8, 1),
	}
}

func (d *remoteDevice) Monitor(influx *utils.Influx, timeLimit time.Time, timeSleep time.Duration) {
	log.Printf("Started goroutine for (device ID: %d)\n", d.ID)

CLIENT_CREATION:
	auth, err := d.auth()
	if err != nil {
		log.Printf("Cannot parse key: %v (device ID: %d)\n", err, d.ID)
		d.ExitSignal <- utils.STATUS_ERROR_KEYFILE
		return
	}

	client, err := d.sshClient(auth)
	if err != nil {
		log.Printf("SSH client error: %v (device ID: %d)\n", err, d.ID)
		d.ExitSignal <- utils.STATUS_ERROR_SSH
		return
	}
	defer client.Close()
	log.Printf("Created SSH client (device ID: %d)\n", d.ID)

	interfaces, err := d.getInterfaces(client)
	if err != nil {
		fmt.Printf("Error with getting interfaces: %v (device ID: %d)\n", err, d.ID)
		d.ExitSignal <- utils.STATUS_ERROR_SSH
		return
	}
	log.Printf("Detected %d interface(s) (device ID: %d)\n", len(interfaces), d.ID)

	failedRuns := 0
	for time.Now().Before(timeLimit) {
		err := d.monitorInterfaces(client, interfaces, influx)
		if err != nil {
			log.Printf("Monitoring error(s) (device ID: %d):\n%v", d.ID, err)
			failedRuns += 1
		}

		if failedRuns > FAILED_RUNS_LIMIT {
			log.Printf("Error limit exceeded (%d errors), trying to reconnect (device ID: %d)\n", failedRuns, d.ID)
			if err := client.Close(); err != nil {
				log.Printf("Closing session error: %v (device ID: %d)\n", err, d.ID)
				d.ExitSignal <- utils.STATUS_ERROR_SSH
				return
			}

			goto CLIENT_CREATION
		}

		time.Sleep(timeSleep)
	}

	if failedRuns != 0 {
		d.ExitSignal <- utils.STATUS_WARNING
	}

	d.ExitSignal <- utils.STATUS_OK
}

func (d *remoteDevice) auth() ([]ssh.AuthMethod, error) {
	if d.Key == nil {
		return []ssh.AuthMethod{ssh.Password(d.Password)}, nil
	}

	signer, err := ssh.ParsePrivateKey(d.Key)
	if err != nil {
		log.Printf("Unable to parse private key: %v (device ID: %d)\n", err, d.ID)
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

	client, err := ssh.Dial("tcp", d.IP+":22", sshCfg)
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
			err = errors.Join(err, fmt.Errorf("%v (interface: %s)\n", err2, inf))
			continue
		}
		defer session.Close()

		got, err2 := session.CombinedOutput(fmt.Sprintf(CMD_SHOW_EEPROM, inf))
		if err2 != nil {
			fmt.Println("error2, interface: ", inf)
			err = errors.Join(err, fmt.Errorf("%v (interface: %s)\n", err2, inf))
			continue
		}

		interfaceData, err2 := d.processData(got)
		if err2 != nil {
			fmt.Println("error3, interface: ", inf)
			err = errors.Join(err, fmt.Errorf("%v (interface: %s)\n", err2, inf))
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
