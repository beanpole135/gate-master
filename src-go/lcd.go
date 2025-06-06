//go:build !arm64

package main

import (
	"fmt"
)

type LCDConfig struct {
	//Config file variables
	Bus_num        int    `json:"i2c_bus_number"`
	Backlight_secs int    `json:"backlight_seconds"`
	Hex_addr       string `json:"hex_address"`
}

func (L *LCDConfig) Setup() (err error) {
	return
}

func (L *LCDConfig) Display(text string) {
	fmt.Println("Putting Text on LCD Display:", text)
}

func (L *LCDConfig) Clear() {

}

func (L *LCDConfig) Close() {

}
