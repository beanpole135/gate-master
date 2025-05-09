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
	step := "Pin Create"
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error opening gate:", step+":", r)
			err = fmt.Errorf("Error opening gate")
		}
	}()

	pin := rpio.Pin(gc.GpioPin)
	step = "PinMode"
	rpio.PinMode(pin, rpio.Input) //Need to ensure the pin is an input first (we input to the pin)
	//Gate is triggered to open by a 1 second state change on the pin
	step = "PullUp"
	pin.PullUp()
	time.Sleep(time.Second) //wait one second
	step = "PullDown"
	pin.PullDown()
	rpio.Close()
	return nil
}
