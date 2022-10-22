package main

// sshConfig := &ssh.ClientConfig{
// 	User:            "",
// 	Auth:            []ssh.AuthMethod{ssh.Password("")},
// 	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
// }

// client, err := ssh.Dial("tcp", "", sshConfig)
// if err != nil {
// 	log.Fatalln(err)
// }

// session, err := client.NewSession()
// if err != nil {
// 	client.Close()
// 	log.Fatalln(err)
// }

// out, err := session.CombinedOutput("ls -l /tmp/")
// if err != nil {
// 	log.Fatalln(err)
// }

// fmt.Println(string(out))
// client.Close()
