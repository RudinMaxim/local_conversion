package main

import (
	"fmt"

	"github.com/RudinMaxim/local_conversion/cmd"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config/")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
	}

	cmd.Execute()
}
