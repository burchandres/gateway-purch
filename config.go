package main

import (
	"log/slog"

	"github.com/spf13/viper"
)

type GatewayConfig struct {
	ServerAddress string `mapstructure:"server-address"`
	TargetRoutes []Route `mapstructure:"target-routes"`
}

type Route struct {
	Name    string `mapstructure:"name"`
	Root    string `mapstructure:"root"`
	Address string `mapstructure:"address"`
}

func ReadConfig() *GatewayConfig {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		slog.Error("error while reading in config.", "error", err.Error())
		panic(err)
	}

	config := &GatewayConfig{}

	if err := viper.Unmarshal(config); err != nil {
		slog.Error("unable to unmarshal config into GatewayConfig struct.", "error", err.Error())
		panic(err)
	}

	return config
}