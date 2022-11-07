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
const SHOW_FIBER_INTERFACES_CMD string = "show-fiber-interfaces"
const EEPROM_DECODER_TYPE int = 0

func UnifyEeprom(device *common.Device, input []byte) []byte {
	if EEPROM_DECODER_TYPE == 0 {
		temp := []byte{}
		for i := 0; i < len(input); i += 33 {
			temp = append(temp, input[i:i+32]...)
		}
		output, err := hex.DecodeString(string(temp))
		if err != nil {
			log.Printf("Error while unifying EEPROM on device with ID: %d\n", device.Id)
			return input
		}

		return output
	} else {
		log.Println("Unknown EEPROM format, handling as unformatted")
		return input
	}
}

func ProcessData(device *common.Device, eeprom []byte) common.InterfaceData {
	eeprom = UnifyEeprom(device, eeprom)
	return common.InterfaceData{
		Temperature: GetTemperature(eeprom),
		Voltage:     GetVoltage(eeprom),
		TxPower:     GetTxPower(eeprom),
		RxPower:     GetRxPower(eeprom),
		Osnr:        GetOsnr(eeprom),
	}
}

func CreateSshClient(device *common.Device) (session *ssh.Client, err error) {
	sshConfig := &ssh.ClientConfig{
		User:            device.Login,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	if signer, err := ssh.ParsePrivateKey(device.Key); err != nil || device.Key == nil || device.Status == 2 {
		if err != nil && device.Key != nil {
			log.Printf("Unable to parse private key for device with ID: %d (%v)\n", device.Id, err)
			device.Status = 2
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(device.Password)}
	} else {
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}

	client, err := ssh.Dial("tcp", device.Ip+":22", sshConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func MonitorData(device *common.Device, influxConfig *common.InfluxConfig, inputSig <-chan bool, outputSig chan<- bool, timeSleep int) {
	log.Printf("Started goroutine for device with ID: %d\n", device.Id)

CLIENT_CREATION:
	interfaces := []string{}

	log.Printf("Creating SSH client for device with ID: %d\n", device.Id)
	client, errStart := CreateSshClient(device)
	if errStart != nil {
		log.Printf("SSH client error on device with ID: %d (%s)\n", device.Id, errStart)
		outputSig <- true
	} else {
		defer client.Close()

		log.Printf("Creating SSH session for interfaces on device with ID: %d\n", device.Id)
		session, errStart := client.NewSession()
		if errStart != nil {
			log.Printf("SSH session error on device with ID: %d (%s)\n", device.Id, errStart)
			outputSig <- true
		} else {
			defer session.Close()

			log.Printf("Getting interfaces on device with ID: %d\n", device.Id)
			byteOutput, errStart := session.CombinedOutput(SHOW_FIBER_INTERFACES_CMD)
			if errStart != nil {
				log.Printf("Error with getting interfaces on device with ID: %d (%s)\n", device.Id, errStart)
				outputSig <- true
			} else {
				interfaces = strings.SplitN(string(byteOutput), "\n", -1)
				log.Printf("Detected %d interfaces on device with ID: %d\n", len(interfaces)-1, device.Id)
			}
		}
	}

	sessionErrors := 0
	for {
		select {
		case <-inputSig:
			log.Printf("Received stop signal on device with ID %d\n", device.Id)
			return
		default:
			for i := 0; i < len(interfaces)-1; i++ {
				session, err := client.NewSession()
				if err != nil {
					log.Printf("SSH session error on device with ID: %d (%s)\n", device.Id, err)
					if sessionErrors > 2 {
						log.Printf("Reconnecting to device with ID: %d (too many session errors)\n", device.Id)
						client.Close()
						goto CLIENT_CREATION
					} else {
						log.Printf("Skiping interface %s on device with ID: %d (session error)\n", interfaces[i], device.Id)
						sessionErrors += 1
					}
					continue
				}
				defer session.Close()

				byteOutput, err := session.CombinedOutput(fmt.Sprintf(SHOW_EEPROM_CMD, interfaces[i]))
				if err != nil {
					log.Printf("Error with interface %s on device with ID: %d (%s)\n", interfaces[i], device.Id, err)
					continue
				}

				interfaceData := ProcessData(device, byteOutput)
				common.InsertToInflux(influxConfig, device.Hostname, interfaces[i], &interfaceData)
			}

			time.Sleep(time.Duration(timeSleep) * time.Second)
		}
	}
}
