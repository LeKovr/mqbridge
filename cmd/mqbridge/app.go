package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/LeKovr/mqbridge"
)

// Config holds all config vars
type Config struct {
	Flags
	Bridge mqbridge.Config `group:"MQ Bridge Options"`
}

// Actual version value will be set at build time
var version = "0.0-dev"

// Run app and exit via given exitFunc
func Run(exitFunc func(code int)) {
	cfg, err := SetupConfig()
	log := SetupLog(err != nil || cfg.Debug)
	defer func() { Shutdown(exitFunc, err, log) }()
	log.Info("mqbridge. Stream messages from PG/NATS/File channel to another PG/NATS/File channel.", "v", version)
	if err != nil || cfg.Version {
		return
	}
	var mqbr *mqbridge.Service
	mqbr, err = mqbridge.New(log, &cfg.Bridge)
	if err != nil {
		return
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	log.Info("Server starting")
	err = mqbr.Run(quit)
	log.Info("Server stopped")
}
