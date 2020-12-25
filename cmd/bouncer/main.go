package main

import (
	"fmt"
	"os"
	"time"

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

	// verify config parsing before proceeding with further app development
	fmt.Printf("Config: %+v\n", cfg)

	log.WithFields(log.Fields{"highPin": cfg.HighPin, "lowPin": cfg.LowPin}).Info("initializing GPIO")
	relay, gpioErr := gpio.Init(cfg.HighPin, cfg.LowPin)
	if gpioErr != nil {
		log.WithFields(log.Fields{"err": gpioErr}).Error("could not initialize GPIO")
		os.Exit(3)
	}

	log.Info("sleeping for 5 seconds")
	time.Sleep(5 * time.Second)
	log.WithFields(log.Fields{"duration": cfg.BounceDuration}).Info("bouncing relay")
	relay.Trigger(cfg.BounceDuration)
}
