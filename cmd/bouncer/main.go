package main

import (
	"fmt"
	"os"

	"github.com/frick/bouncer/pkg/conf"
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
	cfg, err := conf.Load()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("could not load configuration")
		os.Exit(127)
	}

	// if -version is passed, print our version and exit
	if cfg.Version {
		fmt.Println(Version)
		os.Exit(0)
	}

	// verify config parsing before proceeding with further app development
	fmt.Printf("Config: %+v\n", cfg)
}
