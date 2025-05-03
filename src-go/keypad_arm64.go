//go:build arm64

package main

import (
	"fmt"
	"time"

	"github.com/warthog618/go-gpiocdev"
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
	lines     []*gpiocdev.Line `json:"-"`
	pressed   int              `json:"-"` //count of lines pressed right now
	up        map[int]bool     `json:"-"`
	pin_cache string           `json:"-"` //current PIN code pending
}

func (K *Keypad) StartWatching() {
	offsets := []int{K.R1, K.R2, K.R3, K.R4, K.C1, K.C2, K.C3}
	K.up = make(map[int]bool)
	for _, offset := range offsets {
		K.up[offset] = false
		l, _ := gpiocdev.RequestLine(K.Chipname, offset, gpiocdev.WithEventHandler(K.handler), gpiocdev.WithBothEdges)
		K.lines = append(K.lines, l)
	}
}

func (K *Keypad) Close() {
	for _, line := range K.lines {
		line.Close()
	}
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

func (K *Keypad) handler(evt gpiocdev.LineEvent) {
	// When a pair of row/col lines are "pressed"
	// - that indicates a particular button itself was pressed
	fmt.Println("Got Keypress with offset:", evt.Offset, "RisingEdge:", evt.Type == gpiocdev.LineEventRisingEdge)
	if evt.Type == gpiocdev.LineEventRisingEdge {
		//Mark this row/column as "pressed"
		K.up[evt.Offset] = true
		K.pressed += 1
	} else {
		//Remove this row/column from "pressed"
		K.up[evt.Offset] = false
		K.pressed -= 1
		return //Stop processing here for releases of keys
	}
	if K.pressed != 2 {
		return
	}
	//Now look for all the row/column pairs and see what is there
	if K.Pressed(K.R1) {
		if K.Pressed(K.C1) {
			K.NumPressed(1)
		} else if K.Pressed(K.C2) {
			K.NumPressed(2)
		} else if K.Pressed(K.C3) {
			K.NumPressed(3)
		}
	} else if K.Pressed(K.R2) {
		if K.Pressed(K.C1) {
			K.NumPressed(4)
		} else if K.Pressed(K.C2) {
			K.NumPressed(5)
		} else if K.Pressed(K.C3) {
			K.NumPressed(6)
		}
	} else if K.Pressed(K.R3) {
		if K.Pressed(K.C1) {
			K.NumPressed(7)
		} else if K.Pressed(K.C2) {
			K.NumPressed(8)
		} else if K.Pressed(K.C3) {
			K.NumPressed(9)
		}
	} else if K.Pressed(K.R4) {
		if K.Pressed(K.C1) {
			K.ClearPressed() // * key
		} else if K.Pressed(K.C2) {
			K.NumPressed(1)
		} else if K.Pressed(K.C3) {
			K.EnterPressed() // # key
		}
	}
}

func (K *Keypad) Pressed(offset int) bool {
	v, ok := K.up[offset]
	if !ok || v == false {
		return false
	}
	return true
}
