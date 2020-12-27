package gpio

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stianeikeland/go-rpio/v4"
)

// GPIO tracks our high and low pins
type GPIO struct {
	IsARM   bool
	HighPin rpio.Pin
	LowPin  rpio.Pin
}

// InitNotARM returns a basic GPIO struct
func InitNotARM(highPinNum, lowPinNum int) (*GPIO, error) {
	log.Warning("not running on ARM, skipping GPIO initialization")
	return &GPIO{
		IsARM: false,
	}, nil
}

// InitARM takes our high and low pin numbers and returns a GPIO struct after initializing their states
func InitARM(highPinNum, lowPinNum int) (*GPIO, error) {
	log.WithFields(log.Fields{"highPin": highPinNum, "lowPin": lowPinNum}).Info("initializing GPIO")

	err := rpio.Open()
	if err != nil {
		return nil, err
	}

	var lowPin rpio.Pin
	if lowPinNum >= 0 {
		lowPin = rpio.Pin(lowPinNum)
		lowPin.Output()
		lowPin.Low()
	}

	highPin := rpio.Pin(highPinNum)
	highPin.Output()
	highPin.Low()

	return &GPIO{
		IsARM:   true,
		LowPin:  lowPin,
		HighPin: highPin,
	}, nil
}

// Trigger takes a duration and actually sets our GPIO "high" pin high for said duration before returning to low
func (g *GPIO) Trigger(duration time.Duration) {
	if g.IsARM {
		log.WithFields(log.Fields{"duration": duration}).Info("triggering relay")
		g.HighPin.High()
		time.Sleep(duration)
		g.HighPin.Low()
	} else {
		log.Info("would have triggered relay, but not running on a Raspberry Pi")
	}
}

// Close unmaps GPIO memory
func (g *GPIO) Close() error {
	if g.IsARM {
		return rpio.Close()
	}
	return nil
}
