package gpio

import (
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

type GPIO struct {
	HighPinNum int
	LowPinNum  int
	HighPin    rpio.Pin
	LowPin     rpio.Pin
}

func Init(highPinNum, lowPinNum int) (*GPIO, error) {
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
		HighPinNum: highPinNum,
		LowPinNum:  lowPinNum,
		LowPin:     lowPin,
		HighPin:    highPin,
	}, nil
}

func (g *GPIO) Trigger(duration time.Duration) {
	g.HighPin.High()
	time.Sleep(duration)
	g.HighPin.Low()
}
