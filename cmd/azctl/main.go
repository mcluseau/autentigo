package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mcluseau/autorizo/client"
)

var (
	az *client.Client

	termIn  = bufio.NewReader(os.Stdin)
	termOut = os.Stderr
)

func main() {
	defaultServer := os.Getenv("AZCTL_SERVER")
	if len(defaultServer) == 0 {
		defaultServer = "http://localhost:8080"
	}

	serverURL := flag.String("server", defaultServer, "Autorizo server URL")
	flag.Parse()

	az = client.New(*serverURL)

	if len(flag.Args()) < 1 {
		fail(errors.New("need a command"))
	}

	// handle termination signals
	sig := make(chan os.Signal, 1)
	go func() {
		<-sig
		resetTerm()
		os.Exit(1)
	}()

	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	// execute command
	switch v := flag.Args()[0]; v {
	case "login":
		login()

	default:
		fail(fmt.Errorf("unknown command: %s", v))
	}
}

func login() {
	termOut.WriteString("username: ")
	username, err := termIn.ReadString('\n')
	fail(err)

	termOut.WriteString("password: \x1b[8m")
	password, err := termIn.ReadString('\n')
	fail(err)

	// remove trailing \n
	username = username[0 : len(username)-1]
	password = password[0 : len(password)-1]

	resetTerm()

	res, err := az.Login(username, password)
	fail(err)

	fmt.Println(res.Token)
}

func fail(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		resetTerm()
		os.Exit(1)
	}
}

func resetTerm() {
	termOut.WriteString("\x1b[0m")
}
