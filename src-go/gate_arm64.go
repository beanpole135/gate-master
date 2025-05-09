//go:build arm64

package main

import (
	"fmt"
	"time"

	"github.com/stianeikeland/go-rpio"
)

type GateConfig struct {
	GpioPin uint8 `json:"gpio_num"`
}

func (gc *GateConfig) OpenGate() (err error) {
	if gc == nil || gc.GpioPin < 1 {
		fmt.Println("No Gate configured!!")
		return fmt.Errorf("No Gate Configured")
	}
	//Setup crash handling for error detection. rpio library panics on errors
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error opening gate:", r)
			err = fmt.Errorf("Error opening gate")
		}
	}()

	pin := rpio.Pin(gc.GpioPin)
	rpio.PinMode(pin, rpio.Output) //Need to toggle the pin as an output first
	//Gate is triggered to open by a 1 second state change on the pin
	pin.PullUp()
	time.Sleep(time.Second) //wait one second
	pin.PullDown()
	rpio.Close()
	return nil
}
