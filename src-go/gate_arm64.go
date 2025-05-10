//go:build arm64

package main

import (
	"fmt"
	"time"
)

type GateConfig struct {
	GpioPin uint8 `json:"gpio_num"`
}

func (gc *GateConfig) OpenGate() (err error) {
	if gc == nil || gc.GpioPin < 1 {
		fmt.Println("No Gate configured!!")
		return fmt.Errorf("No Gate Configured")
	}
	err = SetPinUp(uint32(gc.GpioPin))
	if err != nil {
		return err
	}
	time.Sleep(time.Second) //wait one second
	err = SetPinDown(uint32(gc.GpioPin))
	if err != nil {
		return err
	}
	return nil
}
