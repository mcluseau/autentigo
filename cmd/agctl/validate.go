package main

import (
	"errors"

	"github.com/spf13/cobra"
)

func init() {
	cmdAzctl.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: "validate a token",
		Run:   doValidate,
	})
}

func doValidate(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fail(errors.New("validate command needs a token"))
	}

	validate(args[0])
}
