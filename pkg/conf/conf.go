package conf

import (
	"context"
	"time"

	"github.com/vimeo/dials/ez"
)

// Config is the configuration struct for bouncer
type Config struct {
	Config         string        `dialsdesc:"path to YAML config"`
	CheckInterval  time.Duration `yaml:"check-interval" dialsdesc:"duration between checks"`
	CheckJitter    time.Duration `yaml:"check-jitter" dialsdesc:"amount of time, plus or minus, by which to skew the check interval"`
	CheckTimeout   time.Duration `yaml:"check-timeout" dialsdesc:"the total time allowed for a check to succeed"`
	RetryInterval  time.Duration `yaml:"retry-interval" dialsdesc:"after a check has failed, duration between retries"`
	RetryJitter    time.Duration `yaml:"retry-jitter" dialsdesc:"amount of time, plus or minus, by which to skew retries"`
	Failures       int           `yaml:"failures" dialsdesc:"number of failures before triggering a 'bounce'"`
	BounceDuration time.Duration `yaml:"bounce-duration" dialsdesc:"how long the bouncer should trigger, ie. off for 10s"`
	BounceTimeout  time.Duration `yaml:"bounce-timeout" dialsdesc:"how long to wait after triggering a bounce to resume the normal check interval"`
	HighPin        int           `yaml:"high-pin" dialsdesc:"GPIO pin (not board pinout) that will be triggered high/3.3V"`
	LowPin         int           `yaml:"low-pin" dialsdesc:"optional: GPIO pin (not board pinout) that is to be used for a low/ground signal"`
	Sites          []string      `yaml:"sites" dialsdesc:"website(s) to check"`
	Version        bool          `dialsdesc:"print version and exit"`
	Debug          bool          `dialsdesc:"enable debug-level logging"`
}

// ConfigPath returns the path to the config file that Dials should read. This
// is particularly helpful when it's desirable to specify the file's path via
// environment variables or command line flags. Dials will first populate the
// configuration struct from environment variables and command line flags
// and then read the config file that the ConfigPath() method returns
func (c *Config) ConfigPath() (string, bool) {
	if c.Config == "" {
		return "", false
	}
	return c.Config, true
}

// Load returns a populated Config struct. In order of increasing precedence:
// YAML config file, environment variables, command-line flags, defaults specified below
func Load() (*Config, error) {
	c := &Config{
		CheckInterval:  time.Duration(2) * time.Minute,
		CheckJitter:    time.Duration(10) * time.Second,
		CheckTimeout:   time.Duration(30) * time.Second,
		RetryInterval:  time.Duration(20) * time.Second,
		RetryJitter:    time.Duration(4) * time.Second,
		Failures:       5,
		BounceDuration: time.Duration(10) * time.Second,
		BounceTimeout:  time.Duration(10) * time.Minute,
		HighPin:        21,
		LowPin:         -1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The following function will populate the config struct by reading the
	// config files, environment variables, and command line flags (order matches
	// the function name) with increasing precedence.
	d, dialsErr := ez.YAMLConfigEnvFlag(ctx, c)
	if dialsErr != nil {
		return c, dialsErr
	}

	// Fill will deep copy the fully-stacked configuration into its argument.
	d.Fill(c)
	return c, nil
}
