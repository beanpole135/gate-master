//go:build !arm64

package main

type Keypad struct{}

func (K *Keypad) StartWatching() {

}

func (K *Keypad) Close() {

}

func (K *Keypad) DisplayOnLCD(text string, seconds int) {

}
