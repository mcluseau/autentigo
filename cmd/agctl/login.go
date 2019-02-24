package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	cmdAzctl.AddCommand(&cobra.Command{
		Use:   "login",
		Short: "login and return the token",
		Run:   doLogin,
	})
}

func doLogin(_ *cobra.Command, _ []string) {
	res, err := az.Login(readUserPass())
	fail(err)

	fmt.Println(res.Token)
}

func readUserPass() (username, password string) {
	termOut.WriteString("username: ")
	username, err := termIn.ReadString('\n')
	fail(err)

	termOut.WriteString("password: \x1b[8m")
	defer resetTerm()
	password, err = termIn.ReadString('\n')
	fail(err)

	// remove trailing \n
	username = username[0 : len(username)-1]
	password = password[0 : len(password)-1]

	return
}
