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
	R1       int    `json:"row1"`
	R2       int    `json:"row2"`
	R3       int    `json:"row3"`
	R4       int    `json:"row4"`
	C1       int    `json:"col1"`
	C2       int    `json:"col2"`
	C3       int    `json:"col3"`
	// Internal variables
	pin_cache string `json:"-"` //current PIN code pending
}

func (K *Keypad) StartWatching() {
	go ScanInputEvents(K.CheckEvents)
}

func (K *Keypad) CheckEvents(diff map[int]PinState) {
	if v, ok := diff[K.R1]; ok && v == PIN_UP {
		if v, ok := diff[K.C1]; ok && v == PIN_UP {
			K.NumPressed(1)
		} else if v, ok := diff[K.C2]; ok && v == PIN_UP {
			K.NumPressed(2)
		} else if v, ok := diff[K.C3]; ok && v == PIN_UP {
			K.NumPressed(3)
		}
	} else if v, ok := diff[K.R2]; ok && v == PIN_UP {
		if v, ok := diff[K.C1]; ok && v == PIN_UP {
			K.NumPressed(4)
		} else if v, ok := diff[K.C2]; ok && v == PIN_UP {
			K.NumPressed(5)
		} else if v, ok := diff[K.C3]; ok && v == PIN_UP {
			K.NumPressed(6)
		}
	} else if v, ok := diff[K.R3]; ok && v == PIN_UP {
		if v, ok := diff[K.C1]; ok && v == PIN_UP {
			K.NumPressed(7)
		} else if v, ok := diff[K.C2]; ok && v == PIN_UP {
			K.NumPressed(8)
		} else if v, ok := diff[K.C3]; ok && v == PIN_UP {
			K.NumPressed(9)
		}
	} else if v, ok := diff[K.R4]; ok && v == PIN_UP {
		if v, ok := diff[K.C1]; ok && v == PIN_UP {
			K.ClearPressed() // * key
		} else if v, ok := diff[K.C2]; ok && v == PIN_UP {
			K.NumPressed(1)
		} else if v, ok := diff[K.C3]; ok && v == PIN_UP {
			K.EnterPressed() // # key
		}
	}
}

func (K *Keypad) Close() {

}

func (K *Keypad) NumPressed(num int) {
	fmt.Println("Number Pressed:", num)
	K.pin_cache += fmt.Sprintf("%d", num)
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
	fmt.Println("Enter Pressed")
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
	fmt.Println("Clear Pressed")
	K.pin_cache = ""
	go ClearLCD(0)
}
