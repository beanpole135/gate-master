//go:build !arm
package main

import (
)

type Keypad struct {}

func (K *Keypad) StartWatching(chipname string) {

}

func (K *Keypad) Close() {

}
