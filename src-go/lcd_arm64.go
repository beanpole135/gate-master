//go:build arm64

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/davecheney/i2c"
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
	//Internal variables
	internal_i2c *i2c.I2C    `json:"-"`
	internal_lcd *i2c.Lcd    `json:"-"`
	bltimer      *time.Timer `json:"-"`
}

func (L *LCDConfig) Setup() (err error) {
	haddr, err := strconv.ParseUint(strings.TrimPrefix(L.Hex_addr, "0x"), 16, 8)
	if err != nil {
		return err
	}
	L.internal_i2c, err = i2c.New(uint8(haddr), L.Bus_num)
	if err != nil {
		return err
	}
	//Config Pins input order: [en, rw, rs, d4, d5, d6, d7, backlight]
	L.internal_lcd, err = i2c.NewLcd(L.internal_i2c, L.Pins.En, L.Pins.Rw, L.Pins.Rs, L.Pins.D4, L.Pins.D5, L.Pins.D6, L.Pins.D7, L.Pins.Backlight)
	//PutW it in "standby" mode initially
	L.internal_lcd.Clear()
	L.internal_lcd.BacklightOff()
	//Setup the backlight timer to turn off after a period of seconds
	L.bltimer = time.NewTimer(time.Duration(L.Backlight_secs) * time.Second)
	go func() {
		for {
			<-L.bltimer.C
			L.internal_lcd.BacklightOff()
		}
	}()
	L.bltimer.Stop() //Don't need to start initially - already off
	return err
}

func (L *LCDConfig) Display(text string) {
	fmt.Println("Display on LCD:", text)
	//Put the text on the screen
	L.internal_lcd.Clear()
	L.internal_lcd.SetPosition(1, 0) //line 1, character 0
	num, err := L.internal_lcd.Write([]byte(text))
	if num < 1 || err != nil {
		fmt.Println("Error writing to LCD: bytes written:", num, ", Error:", err)
	}

	//Turn on the backlight since something changed on the screen
	L.internal_lcd.BacklightOn()
	L.bltimer.Reset(time.Duration(L.Backlight_secs) * time.Second)
}

func (L *LCDConfig) Clear() {
	fmt.Println("Clearing LCD")
	L.internal_lcd.Clear()
}
