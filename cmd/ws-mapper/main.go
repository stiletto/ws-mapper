package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/stiletto/ws-mapper/wsmapper"
	"gopkg.in/yaml.v3"
)

func main() {
	config := wsmapper.DefaultConfig()
	var configFile string
	var example bool
	var configTemplate bool
	defaultConfig := "/etc/ws-mapper.yaml"
	flag.BoolVar(&example, "example", false, "Print example config")
	flag.StringVar(&configFile, "config", defaultConfig, "Config file")
	flag.BoolVar(&configTemplate, "config-template", false, "Print active config and exit")
	flag.Parse()
	if example {
		config = wsmapper.ExampleConfig()
		yaml.NewEncoder(os.Stdout).Encode(config)
		return
	}
	if f, err := os.Open(configFile); err == nil {
		err = yaml.NewDecoder(f).Decode(&config)
		f.Close()
		if err != nil {
			slog.Error("Failed to read config file", "err", err, "config", configFile)
			os.Exit(78)
			return
		}
	} else {
		slog.Error("Failed to open config file", "err", err, "config", configFile)
		if !configTemplate {
			os.Exit(78)
			return
		}
	}
	if configTemplate {
		yaml.NewEncoder(os.Stdout).Encode(config)
		return
	} else {
		cfgBytes, _ := yaml.Marshal(config)
		slog.Info("Active config:\n" + string(cfgBytes))
	}
	if err := wsmapper.CheckAndFixConfig(&config); err != nil {
		slog.Error("Configuration error", "err", err, "config", configFile)
		os.Exit(78)
		return
	}
	wsm, err := wsmapper.NewWSMapper(&config, slog.Default())
	if err != nil {
		slog.Error("Failed to create wsmapper", "err", err)
		os.Exit(1)
		return
	}
	err = wsm.Serve()
	if err != nil {
		slog.Error("Failed to serve", "err", err)
	}
}
