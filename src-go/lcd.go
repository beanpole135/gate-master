//go:build !arm64

package main

import (
	"fmt"
)

type LCDConfig struct {
	//Config file variables
	backlight_secs int `json:"backlight_seconds"`
}

func (L *LCDConfig) Setup() (err error) {
	return
}

func (L *LCDConfig) Display(text string) {
	fmt.Println("Putting Text on LCD Display:", text)
}

func (L *LCDConfig) Clear() {

}
