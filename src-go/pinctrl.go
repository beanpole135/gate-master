package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type PinState int

const (
	PIN_UNKNOWN PinState = iota
	PIN_UP
	PIN_DOWN
)

func ReadPins(list string) map[int]PinState {
	// Input list: "1,2,4,6" or "1,4-8,20" (comma-separated and ranges, no spaces though)
	//Line format: "<number>: ip    [pu|pd] | [hi|lo] // [Human-Readable name] = [input]"
	// Example: "2: ip    pu | hi // GPIO2 = input"
	out, err := exec.Command("pinctrl", "get", list).Output()
	lines := strings.Split(string(out), "\n")
	state := make(map[int]PinState)
	if err != nil {
		fmt.Println("Got error [pinctrl get]:", err)
		return state
	}
	for _, line := range lines {
		if list == "" && !strings.HasSuffix(line, "input") {
			continue //skip non-input pins
		}
		words := strings.Fields(line)
		pin := strings.TrimSuffix(words[0], ":")
		pnum, err := strconv.Atoi(pin)
		if err != nil {
			continue
		}
		switch words[2] {
		case "pu":
			state[pnum] = PIN_UP
		case "pd":
			state[pnum] = PIN_DOWN
		}
	}
	return state
}

/*func GetPinState(pin uint32) (PinState, error) {
	out, err := exec.Command(fmt.Sprintf("pinctrl get %d", pin)).Output()
	if err != nil {
		return PIN_UNKNOWN, err
	}
	return PIN_UNKNOWN, nil
}*/

// Primary "set" functions
func SetPinUp(pin uint32) error {
	err := exec.Command("pinctrl", "set", fmt.Sprintf("%d", pin), "pu").Run()
	if err != nil {
		fmt.Println("Got Error [SetPinUp]", err)
	}
	return err
}

func SetPinDown(pin uint32) error {
	err := exec.Command("pinctrl", "set", fmt.Sprintf("%d", pin), "pd").Run()
	if err != nil {
		fmt.Println("Got Error [SetPinUp]", err)
	}
	return err
}

func SetInput(pin uint32) error {
	err := exec.Command("pinctrl", "set", fmt.Sprintf("%d", pin), "ip").Run()
	if err != nil {
		fmt.Println("Got Error [SetInput]", err)
	}
	return err
}

func SetOutput(pin uint32) error {
	err := exec.Command("pinctrl", "set", fmt.Sprintf("%d", pin), "op").Run()
	if err != nil {
		fmt.Println("Got Error [SetOutput]", err)
	}
	return err
}

func SetOutputDriveHigh(pin uint32) error {
	err := exec.Command("pinctrl", "set", fmt.Sprintf("%d", pin), "dh").Run()
	if err != nil {
		fmt.Println("Got Error [SetOutputDriveHigh]", err)
	}
	return err
}

func SetOutputDriveLow(pin uint32) error {
	err := exec.Command("pinctrl", "set", fmt.Sprintf("%d", pin), "dl").Run()
	if err != nil {
		fmt.Println("Got Error [SetOutputDriveLow]", err)
	}
	return err
}

// Primary "scanning" function for watching for input events
type EventHandler func(map[int]PinState)

func ScanInputEvents(fn EventHandler) {
	//Make sure you start this with "go ScanInputEvents(something)"
	prev := ReadPins("")
	var now map[int]PinState
	var diff map[int]PinState
	for {
		now = ReadPins("")
		diff = make(map[int]PinState)
		//Look for differences
		for k, v := range prev {
			nv, ok := now[k]
			if !ok || nv == v {
				continue //no change detected
			}
			diff[k] = nv //Save the new value into the difference map
		}
		if len(diff) > 0 {
			fmt.Println("Got Diff:", diff) // for debugging
			go fn(diff)
		}
		// Now replace the previous map with the now one and get ready for the next check
		prev = now
		// Small pause to prevent overloading the system
		time.Sleep(100 * time.Millisecond) //10 scans per second maximum
	}
}
