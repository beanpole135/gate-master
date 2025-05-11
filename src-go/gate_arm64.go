//go:build arm64

package main

import (
	"fmt"
	"sync"
	"time"
)

type GateConfig struct {
	GpioPin uint32     `json:"gpio_num"`
	Invert  bool       `json:"invert_drive"`
	locker  sync.Mutex `json:"-"`
}

func (gc *GateConfig) SetupGate() error {
	if gc == nil || gc.GpioPin < 1 {
		fmt.Println("No Gate configured!!")
		return fmt.Errorf("No Gate Configured")
	}
	//Now set the Pin to the right settings
	SetOutput(gc.GpioPin) //Set as output device
	gc.SetDrive(false)
	return nil
}

func (gc *GateConfig) OpenGate() (err error) {
	if gc == nil || gc.GpioPin < 1 {
		fmt.Println("No Gate configured!!")
		return fmt.Errorf("No Gate Configured")
	}
	if !gc.locker.TryLock() {
		// Already locked - no need to trigger this again
		// Makes sure that multiple gate open requests get collapsed to a single trigger
		return nil
	}
	defer gc.locker.Unlock()
	// Turn it on for 1 second, then turn it back off again
	err = gc.SetDrive(true)
	if err != nil {
		return err
	}
	time.Sleep(time.Second) //wait one second
	err = gc.SetDrive(false)
	if err != nil {
		return err
	}
	return nil
}

func (gc *GateConfig) SetDrive(on bool) error {
	if gc.Invert {
		on = !on //reverse the order
	}
	if on {
		return SetOutputDriveHigh(gc.GpioPin)
	} else {
		return SetOutputDriveLow(gc.GpioPin)
	}
}
