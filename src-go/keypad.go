package main

import (
	"fmt"
	"github.com/warthog618/go-gpiocdev"
)

type Key int
const (
	Key_0  Key = iota
	Key_1
	Key_2
	Key_3
	Key_4
	Key_5
	Key_6
	Key_7
	Key_8
	Key_9
	Key_Enter
	Key_Clear
)

// Map of our internal keys to the RPI offset
keymap := make(map[Key]int){
	0: Key_0,
	1: Key_1,
	3: Key_2,
	5: Key_3,
	7: Key_4,
	9: Key_5,
	11: Key_6,
	13: Key_7,
	15: Key_8,
	17: Key_9,
	2: Key_Enter,
	4: Key_Clear,
}


type Keypad struct {
	lines []gpiocdev.Line
}

func (K *Keypad) StartWatching(chipname string) {
	for offset, key := range keymap {
		l, _ = gpiocdev.RequestLine(chipname, offset, gpiocdev.WithEventHandler(K.handler), gpiocdev.WithFallingEdge)
	}
}

func (K *Keypad) Close() {
	for _, line := range K.lines {
		line.Close()
	}
}

func (K *Keypad) NumPressed(num int) {
	fmt.Println("Number Pressed:", num)
}

func (K *Keypad) EnterPressed() {
	fmt.Println("Enter Pressed")
}

func (K *Keypad) ClearPressed() {
	fmt.Println("Clear Pressed")
}

func (K *Keypad) handler(evt gpiocdev.LineEvent) {
	//Fetch the offset and convert back to our internal key enum
	fmt.Println("Got Keypress with offset:", evt.Offset)
	key, ok := keymap[evt.Offset]
	if !ok { 
		return
	}
	switch key {
	case Key_Enter:
		K.EnterPressed()
	case Key_Clear:
		K.ClearPressed()
	case Key_0:
		K.NumPressed(0)
	case Key_1:
		K.NumPressed(1)
	case Key_2:
		K.NumPressed(2)
	case Key_3:
		K.NumPressed(3)
	case Key_4:
		K.NumPressed(4)
	case Key_5:
		K.NumPressed(5)
	case Key_6:
		K.NumPressed(6)
	case Key_7:
		K.NumPressed(7)
	case Key_8:
		K.NumPressed(8)
	case Key_9:
		K.NumPressed(9)
	}
}