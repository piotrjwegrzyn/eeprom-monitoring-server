package cmds

import (
	"encoding/hex"
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
	EEPROM_DECODER_TYPE       uint8  = 0
)

type remoteDevice struct {
	utils.Device
}

func (d *remoteDevice) decodeEeprom(input []byte, decoder uint8) []byte {
	if decoder == 0 {
		temp := []byte{}
		for i := 0; i < len(input); i += 33 {
			temp = append(temp, input[i:i+32]...)
		}

		output, err := hex.DecodeString(string(temp))
		if err != nil {
			log.Printf("Error with decoding EEPROM on device with ID: %d\n", d.ID)
			return input
		}

		return output
	}

	log.Println("Unknown EEPROM format, handling as unformatted")
	return input
}

func (d *remoteDevice) ProcessData(input []byte) utils.InterfaceData {
	decoded := d.decodeEeprom(input, EEPROM_DECODER_TYPE)

	return utils.InterfaceData{
		Temperature: temperature(decoded),
		Voltage:     voltage(decoded),
		TxPower:     txPower(decoded),
		RxPower:     rxPower(decoded),
		Osnr:        osnr(decoded),
	}
}

func (d *remoteDevice) NewSSHClient() (session *ssh.Client, err error) {
	sshCfg := &ssh.ClientConfig{
		Auth:            d.auth(),
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

func (d remoteDevice) auth() []ssh.AuthMethod {
	if d.Key == nil {
		return []ssh.AuthMethod{ssh.Password(d.Password)}
	}

	signer, err := ssh.ParsePrivateKey(d.Key)
	if err != nil {
		log.Printf("Unable to parse private key for device with ID: %d (%v)\n", d.ID, err)
		d.Status = 2

		return []ssh.AuthMethod{ssh.Password(d.Password)}
	}

	return []ssh.AuthMethod{ssh.PublicKeys(signer)}
}

func (d *remoteDevice) MonitorData(influx *utils.Influx, inputSig <-chan bool, outputSig chan<- bool, timeSleep int) {
	log.Printf("Started goroutine for device with ID: %d\n", d.ID)

CLIENT_CREATION:
	interfaces := []string{}

	client, errStart := d.NewSSHClient()
	if errStart != nil {
		log.Printf("SSH client error: %v (device ID: %d)\n", errStart, d.ID)
		outputSig <- true
		return
	}
	defer client.Close()
	log.Printf("Created SSH client (device ID: %d)\n", d.ID)

	session, errStart := client.NewSession()
	if errStart != nil {
		log.Printf("SSH session error: %v (device ID: %d)\n", errStart, d.ID)
		outputSig <- true
		return
	}
	defer session.Close()

	byteOutput, errStart := session.CombinedOutput(CMD_SHOW_FIBER_INTERFACES)
	if errStart != nil {
		log.Printf("Getting interfaces error: %v (device ID: %d)\n", errStart, d.ID)
		outputSig <- true
		return
	}

	if err := client.Close(); err != nil {
		log.Printf("Closing session error: %v (device ID: %d)\n", errStart, d.ID)
		outputSig <- true
		return
	}
	d.Status = 0

	interfaces = strings.SplitN(string(byteOutput), "\n", -1)
	log.Printf("Detected %d interfaces (device ID: %d)\n", len(interfaces)-1, d.ID)

	sessionErrors := 0
	for {
		select {
		case <-inputSig:
			log.Printf("Received stop signal (device ID %d)\n", d.ID)
			return
		default:
			if d.monitor(client, interfaces, &sessionErrors, influx) {
				if err := client.Close(); err != nil {
					log.Printf("Closing session error: %v (device ID: %d)\n", err, d.ID)
					return
				}

				goto CLIENT_CREATION
			}

			time.Sleep(time.Duration(timeSleep) * time.Second)
		}
	}
}

func (d *remoteDevice) monitor(client *ssh.Client, interfaces []string, sessionErrors *int, influx *utils.Influx) (reconnect bool) {
	for i := 0; i < len(interfaces)-1; i++ {
		session, err := client.NewSession()
		if err != nil {
			if *sessionErrors > 2 {
				log.Printf("Too many session errors, reconnecting (device ID: %d)\n", d.ID)
				reconnect = true
				return
			}

			log.Printf("Skiping interface %s due to error: %v (device ID: %d)\n", interfaces[i], err, d.ID)
			*sessionErrors += 1
			continue
		}
		defer session.Close()

		byteOutput, err := session.CombinedOutput(fmt.Sprintf(CMD_SHOW_EEPROM, interfaces[i]))
		if err != nil {
			log.Printf("Skiping interface %s due to error: %v (device ID: %d)\n", interfaces[i], err, d.ID)
			continue
		}

		interfaceData := d.ProcessData(byteOutput)
		influx.Insert(d.Hostname, interfaces[i], &interfaceData)
	}

	return
}
