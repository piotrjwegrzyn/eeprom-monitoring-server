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
)

type remoteDevice struct {
	utils.Device
	decode func([]byte) ([]byte, error)
	Out    chan int8
}

func NewRemoteDevice(dev utils.Device, decode func([]byte) ([]byte, error)) *remoteDevice {
	return &remoteDevice{
		Device: dev,
		decode: decode,
		Out:    make(chan int8, 1),
	}
}

func (d *remoteDevice) Monitor(influx *utils.Influx, timeLimit time.Time, timeSleep time.Duration) {
	log.Printf("Started goroutine for (device ID: %d)\n", d.ID)

CLIENT_CREATION:
	auth, err := d.auth()
	if err != nil {
		log.Printf("Cannot parse key: %v (device ID: %d)\n", err, d.ID)
		d.Out <- utils.STATUS_ERROR_KEYFILE
		return
	}

	client, err := d.sshClient(auth)
	if err != nil {
		log.Printf("SSH client error: %v (device ID: %d)\n", err, d.ID)
		d.Out <- utils.STATUS_ERROR_SSH
		return
	}
	defer client.Close()
	log.Printf("Created SSH client (device ID: %d)\n", d.ID)

	interfaces, err := d.getInterfaces(client)
	if err != nil {
		fmt.Printf("Getting interfaces error: %v (device ID: %d)\n", err, d.ID)
		d.Out <- utils.STATUS_ERROR_SSH
		return
	}
	log.Printf("Detected %d interfaces (device ID: %d)\n", len(interfaces)-1, d.ID)

	for time.Now().Before(timeLimit) {
		failCount, err := d.monitorInterfaces(client, interfaces, influx)
		if err != nil {
			log.Printf("Error(s) occurred during monitoring (device ID: %d):\n%v\n", d.ID, err)
		}

		if failCount > 5 {
			log.Printf("Too much errors (%d), trying to reconnect (device ID: %d)\n", failCount, d.ID)
			if err := client.Close(); err != nil {
				log.Printf("Closing session error: %v (device ID: %d)\n", err, d.ID)
				d.Out <- utils.STATUS_ERROR_SSH
				return
			}

			goto CLIENT_CREATION
		}

		time.Sleep(timeSleep)
	}

	d.Out <- utils.STATUS_OK
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

	byteOutput, err := session.CombinedOutput(CMD_SHOW_FIBER_INTERFACES)
	if err != nil {
		return nil, err
	}

	return strings.SplitN(string(byteOutput), "\n", -1), nil
}

func (d *remoteDevice) monitorInterfaces(client *ssh.Client, interfaces []string, influx *utils.Influx) (failCount int, err error) {
	for i := 0; i < len(interfaces); i++ {
		session, err2 := client.NewSession()
		if err2 != nil {
			errors.Join(err, fmt.Errorf("%v (interface: %s)\n", err, interfaces[i]))
			failCount += 1
			continue
		}
		defer session.Close()

		byteOutput, err2 := session.CombinedOutput(fmt.Sprintf(CMD_SHOW_EEPROM, interfaces[i]))
		if err2 != nil {
			errors.Join(err, fmt.Errorf("%v (interface: %s)\n", err, interfaces[i]))
			failCount += 1
			continue
		}

		interfaceData, err2 := d.processData(byteOutput)
		if err2 != nil {
			errors.Join(err, fmt.Errorf("%v (interface: %s)\n", err, interfaces[i]))
			failCount += 1
			continue
		}

		influx.Insert(d.Hostname, interfaces[i], &interfaceData)
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
