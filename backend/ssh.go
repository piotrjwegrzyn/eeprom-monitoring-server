package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	common "pi-wegrzyn/common"
)

const SHOW_EEPROM_CMD string = "show-eeprom %s"
const SHOW_FIBER_INTERFACES string = "show-fiber-interfaces"
const EEPROM_DECODER int = 0

func UnifyEeprom(device *common.Device, input []byte) []byte {
	if EEPROM_DECODER == 0 {
		temp := []byte{}
		for i := 0; i < len(input); i += 33 {
			temp = append(temp, input[i:i+32]...)
		}
		output, err := hex.DecodeString(string(temp))
		if err != nil {
			log.Printf("Error while unifying EEPROM on device %d\n", device.Id)
			return input
		}

		return output
	} else {
		log.Println("Unknown EEPROM format, handling as unformatted")
		return input
	}
}

func ProcessData(device *common.Device, interfaceName string, eeprom []byte) {
	eeprom = UnifyEeprom(device, eeprom)

	temperature := GetTemperature(eeprom)
	voltage := GetVoltage(eeprom)
	txPwr := GetTxPower(eeprom)
	rxPwr := GetRxPower(eeprom)
	osnr := GetOsnr(eeprom)

	fmt.Printf("Device %d/interface %s: T:%.3f, Vcc:%.3f, TxPwr:%.3f, RxPwr:%.3f, OSNR:%.3f\n", device.Id, interfaceName, temperature, voltage, txPwr, rxPwr, osnr)
}

func CreateSshClient(device *common.Device) (session *ssh.Client, err error) {
	sshConfig := &ssh.ClientConfig{
		User:            device.Login,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if signer, err := ssh.ParsePrivateKey(device.Key); err != nil || device.Key == nil || device.Status == 2 {
		if err != nil && device.Key != nil {
			log.Printf("Unable to parse private key for device %d (%v)", device.Id, err)
			device.Status = 2
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(device.Password)}
	} else {
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
		device.Status = 10
	}

	client, err := ssh.Dial("tcp", device.Ip+":22", sshConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func MonitorData(device *common.Device, inputSig <-chan bool, outputSig chan<- bool, timeSleep int) {
CLIENT_CREATION:
	client, err := CreateSshClient(device)
	if err != nil {
		log.Printf("SSH error on %s (%s)\n", device.Ip, err)
		outputSig <- true
		return
	}

	session, err := client.NewSession()
	if err != nil {
		defer client.Close()
		log.Printf("SSH session error on %s (%s)\n", device.Ip, err)
		outputSig <- true
		return
	}

	byteOutput, err := session.CombinedOutput(SHOW_FIBER_INTERFACES)
	if err != nil {
		log.Printf("Error with getting list of interfaces for device %s (%s)", device.Ip, err)
		outputSig <- true
		return
	}
	session.Close()

	interfaces := strings.SplitN(string(byteOutput), "\n", -1)

	for {
		select {
		case <-inputSig:
			defer client.Close()
			return
		default:
			sessionError := false
			for i := 0; i < len(interfaces)-1; i++ {
				session, err = client.NewSession()
				if err != nil {
					log.Printf("SSH session error on %s (%s)\n", device.Ip, err)
					if !sessionError {
						goto CLIENT_CREATION
					} else {
						sessionError = false
					}
					continue
				}
				defer session.Close()

				byteOutput, err := session.CombinedOutput(fmt.Sprintf(SHOW_EEPROM_CMD, interfaces[i]))
				if err != nil {
					log.Printf("Error with EEPROM from interface %s for device %d (%s)\n", interfaces[i], device.Id, err)
					continue
				}

				ProcessData(device, interfaces[i], byteOutput)
			}

			time.Sleep(time.Duration(timeSleep) * time.Second)
		}
	}
}
