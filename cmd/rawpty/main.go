package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
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

	if err := session.RequestPty(getenv("TERM", "vt100"), 204, 60, nil); err != nil {
		log.Fatalf("cannot get remote pseudo terminal: %v", err)
	}

	// copy environment variables to cmd
	var buffer bytes.Buffer
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		buffer.WriteString(fmt.Sprintf(`%s="%s"`, pair[0], pair[1]))
		buffer.WriteString(" ")
	}
	// the first argument contains a string with the command to execute
	buffer.WriteString(os.Args[1])

	if err := session.Run(buffer.String()); err != nil {
		log.Fatalf("error running command %s", err)
	}
}
