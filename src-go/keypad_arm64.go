//go:build arm64

package main

import (
	"fmt"
	"time"
)

/*
	4x3 Keypad
	 [1, 2, 3]
	 [4, 5, 6]
	 [7, 8, 9]
	 [*, 0, #]

Examples:
R1 + C1 = Key "1"
R2 + C3 = Key "6"
*/
func DisplayOnLCD(text string, seconds int) {
	fmt.Println("[TODO] Display on LCD:", text)
	if seconds > 0 {
		go ClearLCD(seconds)
	}
}

func ClearLCD(seconds int) {
	if seconds > 0 {
		time.Sleep(time.Duration(seconds) * time.Second)
	}
	DisplayOnLCD("", 0)
}

type Keypad struct {
	// Configuration from file
	Chipname string `json:"chipname"`
	R1       uint32 `json:"row1"`
	R2       uint32 `json:"row2"`
	R3       uint32 `json:"row3"`
	R4       uint32 `json:"row4"`
	C1       uint32 `json:"col1"`
	C2       uint32 `json:"col2"`
	C3       uint32 `json:"col3"`
	// Internal variables
	pin_cache string              `json:"-"` //current PIN code pending
	col_scan  string              `json:"-"` //quick scan string so we only assemble once
	key_state map[string]PinState `json:"-"` //so we can de-duplicate events
}

func (K *Keypad) StartWatching() {
	// Rows are outputs and drivers
	// Columns are inputs and what we watch for changes when checking a row
	SetOutput(K.R1)
	SetOutputDriveLow(K.R1)
	SetOutput(K.R2)
	SetOutputDriveLow(K.R2)
	SetOutput(K.R3)
	SetOutputDriveLow(K.R3)
	SetOutput(K.R4)
	SetOutputDriveLow(K.R4)
	SetInput(K.C1)
	SetPinDown(K.C1)
	SetInput(K.C2)
	SetPinDown(K.C2)
	SetInput(K.C3)
	SetPinDown(K.C3)
	K.col_scan = fmt.Sprintf("%d,%d,%d", K.C1, K.C2, K.C3)
	K.key_state = make(map[string]PinState)
	go K.watchKeys()
}

func (K *Keypad) watchKeys() {
	for {
		K.readLine(K.R1, []string{"1", "2", "3"})
		K.readLine(K.R2, []string{"4", "5", "6"})
		K.readLine(K.R3, []string{"7", "8", "9"})
		K.readLine(K.R4, []string{"*", "0", "#"})
		time.Sleep(10 * time.Millisecond)
	}
}

func (K *Keypad) readLine(row uint32, vals []string) {
	//Send a signal through the row
	SetOutputDriveHigh(row)
	//Check states of column inputs
	diff := ReadPins(K.col_scan, true) //Need hi/lo checks
	for pin, state := range diff {
		switch uint32(pin) {
		case K.C1:
			K.KeyPressed(vals[0], state)
		case K.C2:
			K.KeyPressed(vals[1], state)
		case K.C3:
			K.KeyPressed(vals[2], state)
		}
	}
	//Turn off the signal in the row
	SetOutputDriveLow(row)
}

func (K *Keypad) KeyPressed(key string, stat PinState) {
	// Need to de-duplicate current key states (UP = pressed, DOWN = not pressed)
	cur, ok := K.key_state[key]
	if !ok {
		K.key_state[key] = stat //update cache
	} else if cur != stat {
		K.key_state[key] = stat //update cache
		if stat == PIN_UP {
			// Only trigger the routine the first time a key is pressed
			// holding key down will not do anything else
			switch key {
			case "*":
				K.ClearPressed()
			case "#":
				K.EnterPressed()
			default:
				K.NumPressed(key)
			}
		}
	}
}

func (K *Keypad) Close() {

}

func (K *Keypad) NumPressed(num string) {
	K.pin_cache += num
	if len(K.pin_cache) > 10 {
		K.ClearPressed()
		return
	}
	disp := ""
	n := 0
	for n < len(K.pin_cache) {
		disp += "*"
		n++
	}
	DisplayOnLCD(disp, 0)
}

func (K *Keypad) EnterPressed() {
	err := fmt.Errorf("PIN Needed")
	if len(K.pin_cache) >= 4 {
		err = CheckPINAndOpen(K.pin_cache)
	}
	K.pin_cache = ""
	if err != nil {
		DisplayOnLCD(err.Error(), 2)
	}
}

func (K *Keypad) ClearPressed() {
	K.pin_cache = ""
	go ClearLCD(0)
}
