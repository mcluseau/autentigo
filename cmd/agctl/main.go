package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/mcluseau/autentigo/client"
)

var (
	az *client.Client

	termIn  = bufio.NewReader(os.Stdin)
	termOut = os.Stderr
)

func main() {
	log.SetPrefix("azctl: ")
	log.SetFlags(log.Lshortfile)
	fail(cmdAzctl.Execute())
}

func fail(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		resetTerm()
		os.Exit(255)
	}
}

func resetTerm() {
	termOut.WriteString("\x1b[0m")
}

func validate(token string) {
	claims := jwt.MapClaims{}
	ok, err := az.Validate(token, claims)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	} else if ok {
		fmt.Println("Token is valid")
		os.Exit(0)
	} else {
		fmt.Println("Token is NOT valid")
		os.Exit(1)
	}
}
