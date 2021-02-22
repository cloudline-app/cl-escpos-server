package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func configSetup() {
	viper.SetConfigFile("settings.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.SetDefault("server.printer_address", "localhost")
	viper.SetDefault("server.printer_port", 9100)
	viper.SetDefault("server.port", 3000)

	viper.SafeWriteConfigAs("settings.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("error when loading config file: settings", err)
		os.Exit(1)
	}

}
