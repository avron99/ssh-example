package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var (
	host          = ""
	port          = "22"
	user          = ""
	ssh_agent_key = ""
)

func main() {
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		panic(err)
	}
	agentsock := agent.NewClient(sock)
	buf, err := os.ReadFile(ssh_agent_key)
	if err != nil {
		panic(err)
	}

	key, err := ssh.ParseRawPrivateKey(buf)
	if err != nil {
		panic(err)
	}
	err = agentsock.Add(agent.AddedKey{
		PrivateKey:       key,
		ConfirmBeforeUse: true,
		LifetimeSecs:     300,
	})
	if err != nil {
		fmt.Printf("add keyring error: %s", err)
	}

	signers, err := agentsock.Signers()
	if err != nil {
		fmt.Printf("create signers error: %s", err)
	}
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", host+":"+port, sshConfig)
	if err != nil {
		fmt.Println(err)
	}
	session, err := client.NewSession()
	if err != nil {
		fmt.Println(err)
	}
	defer session.Close()
	stdin, err := session.StdinPipe()
	if err != nil {
		fmt.Println(err)
	}
	var b bytes.Buffer
	session.Stdout = &b

	err = session.Shell()
	if err != nil {
		fmt.Println(err)
	}

	commands := []string{
		"ls -la",
		"exit",
	}
	for _, cmd := range commands {
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			fmt.Println(err)
		}
	}

	err = session.Wait()
	if err != nil {
		fmt.Println(err)
	}
	words := strings.Fields(b.String())
	fmt.Println(words)
}
