// socket-askpass answers an SSH_ASKPASS or SUDO_ASKPASS prompt with a secret
// read from a Unix socket.
package main

import (
	"fmt"
	"os"
	"time"
)

const (
	socketEnv     = "SOCKET_ASKPASS_PATH"
	answerTimeout = time.Minute
)

func main() {
	// ssh passes the prompt as an argument, and a prompt can begin with "-".
	// The flag package would take it for a flag and exit
	if len(os.Args) > 1 && os.Args[1] == "-version" {
		printVersion()
		return
	}

	path := os.Getenv(socketEnv)
	if path == "" {
		fmt.Fprintf(os.Stderr, "%s is not set\n", socketEnv)
		os.Exit(1)
	}

	if err := copySecret(os.Stdout, path, answerTimeout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
