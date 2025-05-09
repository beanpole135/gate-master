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

func (gc *GateConfig) OpenGate() {
	if gc == nil || gc.GpioPin < 1 {
		fmt.Println("No Gate configured!!")
		return
	}
	pin := rpio.Pin(gc.GpioPin)
	//Gate is triggered to open by a 1 second state change on the pin
	pin.Toggle()
	time.Sleep(time.Second) //wait one second
	pin.Toggle()
	rpio.Close()
}
