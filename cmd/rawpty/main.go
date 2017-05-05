package main

import (
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func getenv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	config := &ssh.ClientConfig{
		User:            getenv("RAWPTY_USERNAME", "test"),
		Auth:            []ssh.AuthMethod{ssh.Password(getenv("RAWPTY_PASSWORD", "ubuntu"))},
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	host := getenv("RAWPTY_HOST", "localhost")
	port := getenv("RAWPTY_PORT", "22")
	connection, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		log.Fatalf("error dialing %v", err)
	}
	defer connection.Close()

	session, err := connection.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %s", err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.RequestPty(getenv("TERM", "vt100"), 80, 40, nil); err != nil {
		log.Fatalf("cannot get remote pseudo terminal: %v", err)
	}

	tstate, terr := terminal.MakeRaw(0)
	if terr != nil {
		log.Fatalf("cannot put local terminal in raw mode: %v", terr)
	}
	defer terminal.Restore(0, tstate)

	// the first argument contains a string with the command to execute
	if err := session.Run(os.Args[1]); err != nil {
		log.Fatalf("error running command %s", err)
	}
}
