//go:build !arm64

package main

import (
	"fmt"
)

type I2CPins struct {
	En        uint8 `json:"en"`
	Rw        uint8 `json:"rw"`
	Rs        uint8 `json:"rs"`
	D4        uint8 `json:"d4"`
	D5        uint8 `json:"d5"`
	D6        uint8 `json:"d6"`
	D7        uint8 `json:"d7"`
	Backlight uint8 `json:"backlight"`
}

type LCDConfig struct {
	//Config file variables
	Bus_num        int     `json:"i2c_bus_number"`
	Backlight_secs int     `json:"backlight_seconds"`
	Pins           I2CPins `json:"i2c_pins"`
	Hex_addr       string  `json:"hex_address"`
}

func (L *LCDConfig) Setup() (err error) {
	return
}

func (L *LCDConfig) Display(text string) {
	fmt.Println("Putting Text on LCD Display:", text)
}

func (L *LCDConfig) Clear() {

}
