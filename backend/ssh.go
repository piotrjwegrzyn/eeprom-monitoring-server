package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	common "pi-wegrzyn/common"
)

const SHOW_EEPROM_CMD string = "show-eeprom %s"
const SHOW_FIBER_INTERFACES string = "show-fiber-interfaces"

func ProcessData(device *common.Device, interfaceName string, eeprom []byte) {
	// TODO
	fmt.Printf("Device %d with interface %s has EEPROM: %X\n", device.Id, interfaceName, eeprom[0])
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
