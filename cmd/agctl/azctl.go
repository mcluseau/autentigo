package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mcluseau/autentigo/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cmdAzctl = &cobra.Command{
		Use:              "azctl",
		Short:            "azctl is a command line interface for the Autorizo API",
		PersistentPreRun: setupAutorizo,
	}

	serverURL string
)

func init() {
	pflags := cmdAzctl.PersistentFlags()
	pflags.StringVarP(&serverURL, "server", "s", "", "Autorizo server URL")

	viper.SetEnvPrefix("AZCTL")

	viper.BindPFlags(pflags)
	viper.SetDefault("server", "http://localhost:8080")

	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
	})
}

func setupAutorizo(_ *cobra.Command, _ []string) {
	if len(serverURL) == 0 {
		serverURL = viper.GetString("server") // XXX ok but.....
	}

	if len(serverURL) == 0 {
		log.Fatal("No server URL defined")
	}

	az = client.New(serverURL)

	// handle termination signals
	sig := make(chan os.Signal, 1)
	go func() {
		<-sig
		resetTerm()
		os.Exit(1)
	}()

	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
}
