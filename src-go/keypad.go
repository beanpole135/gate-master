//go:build !arm

package main

import ()

type Keypad struct{}

func (K *Keypad) StartWatching() {

}

func (K *Keypad) Close() {

}
