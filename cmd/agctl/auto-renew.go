package main

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	cmdAzctl.AddCommand(&cobra.Command{
		Use:   "auto-renew",
		Short: "call a command before a token expires",
		Run:   doAutoRenew,
	})
}

func doAutoRenew(cmd *cobra.Command, args []string) {
	// TODO
	log.Fatal("not yet implemented")
}
