package main

import (
	"fmt"
	"os"

	"github.com/frick/bouncer/pkg/checks"
	"github.com/frick/bouncer/pkg/conf"
	"github.com/frick/bouncer/pkg/gpio"
	log "github.com/sirupsen/logrus"
)

const (
	// Version is the current version of the software
	Version = "0.1.0"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
}

func main() {
	// load our config
	cfg, configErr := conf.Load()
	if configErr != nil {
		log.WithFields(log.Fields{"err": configErr}).Error("could not load configuration")
		os.Exit(127)
	}

	// if -version is passed, print our version and exit
	if cfg.Version {
		fmt.Println(Version)
		os.Exit(0)
	}

	// enable debug level logging if specified
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	log.WithFields(log.Fields{"app": os.Args[0], "version": Version}).Info("starting")
	defer log.WithFields(log.Fields{"app": os.Args[0], "version": Version}).Info("exiting")
	log.WithFields(log.Fields{"config": fmt.Sprintf("%+v", cfg)}).Debug("final configuration")

	// initialize our GPIO interface to our relay
	relay, gpioErr := gpio.Init(cfg.HighPin, cfg.LowPin)
	if gpioErr != nil {
		log.WithFields(log.Fields{"err": gpioErr}).Error("could not initialize GPIO")
		os.Exit(3)
	}
	defer relay.Close()

	// start our never-ending loop of internet connectivity checking
	checks.CheckLoop(cfg, relay)
}
