//go:build !arm64

package main

import (
	"fmt"
)

type I2CPins struct {
	en        uint8 `json:"en`
	rw        uint8 `json:"rw`
	rs        uint8 `json:"rs`
	d4        uint8 `json:"d4`
	d5        uint8 `json:"d5`
	d6        uint8 `json:"d6`
	d7        uint8 `json:"d7`
	backlight uint8 `json:"backlight`
}

type LCDConfig struct {
	//Config file variables
	bus_num        int     `json:"i2c_bus_number"`
	backlight_secs int     `json:"backlight_seconds"`
	pins           I2CPins `json:"i2c_pins"`
}

func (L *LCDConfig) Setup() (err error) {
	return
}

func (L *LCDConfig) Display(text string) {
	fmt.Println("Putting Text on LCD Display:", text)
}

func (L *LCDConfig) Clear() {

}
