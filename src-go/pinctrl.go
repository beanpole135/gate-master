package main

import (
	"fmt"
	"os/exec"
)

type PinState int

const (
	PIN_UNKNOWN PinState = iota
	PIN_UP
	PIN_DOWN
)

func ReadGetLines(output string) map[int]PinState {
	//Line format:
	//out, err := exec.Command("pinctrl get")
	return nil
}

func GetPinState(uint32 pin) (PinState, error) {
	out, err := exec.Command(fmt.Sprintf("pinctrl get $d", pin)).Output()
	if err != nil {
		return PIN_UNKNOWN, err
	}
	return PIN_UNKNOWN, nil
}
