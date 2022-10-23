package main

import (
	"log"
	"time"

	common "pi-wegrzyn/common"

	"golang.org/x/crypto/ssh"
)

func MonitorData(device *common.Device, signal chan int, timeSleep int) error {
	sshConfig := &ssh.ClientConfig{
		User:            device.Login,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if signer, err := ssh.ParsePrivateKey(device.Key); err != nil || device.Key == nil || device.Status == 2 {
		if err != nil {
			log.Printf("Unable to parse private key for device %d (%v)", device.Id, err)
			device.Status = 2
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(device.Password)}

	} else {
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}

	client, err := ssh.Dial("tcp", device.Ip+":22", sshConfig)
	if err != nil {
		log.Printf("Error while dialing %s (%s)\n", device.Ip, err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		log.Printf("Error with establishing connection with %s (%s)\n", device.Ip, err)
	}

	for {
		select {
		case <-signal:
			defer session.Close()
			defer client.Close()
			return nil
		default:

			out, err := session.CombinedOutput("ls -l /tmp/ | head -1")
			if err != nil {
				log.Printf("Error with getting output from %s (%s)\n", device.Ip, err)
			}

			log.Println(string(out))

			time.Sleep(time.Duration(timeSleep) * time.Second)
		}
	}
}
