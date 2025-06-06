//go:build arm64

package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	hd44780 "github.com/d2r2/go-hd44780"
	i2c "github.com/d2r2/go-i2c"
)

type LCDConfig struct {
	//Config file variables
	Bus_num        int    `json:"i2c_bus_number"`
	Backlight_secs int    `json:"backlight_seconds"`
	Hex_addr       string `json:"hex_address"`
	//Internal variables
	internal_i2c *i2c.I2C     `json:"-"`
	internal_lcd *hd44780.Lcd `json:"-"`
	bltimer      *time.Timer  `json:"-"`
}

func (L *LCDConfig) Setup() (err error) {
	haddr, err := strconv.ParseUint(strings.TrimPrefix(L.Hex_addr, "0x"), 16, 8)
	if err != nil {
		return err
	}
	// Create a new I2C bus connection.
	L.internal_i2c, err = i2c.NewI2C(uint8(haddr), L.Bus_num)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new LCD instance.
	// We specify the I2C connection, and the LCD's dimensions (16 columns, 2 rows).
	L.internal_lcd, err = hd44780.NewLcd(L.internal_i2c, hd44780.LCD_16x2) // The library uses 16x2 for 1x16 displays as well.
	if err != nil {
		log.Fatal(err)
	}

	//Put it in "standby" mode initially
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
	//fmt.Println("Display on LCD:", text)
	//Put the text on the screen
	L.internal_lcd.Clear()
	err := L.internal_lcd.ShowMessage(text, hd44780.SHOW_LINE_1)
	if err != nil {
		fmt.Println("Error writing to LCD: bytes written:", err)
	}

	//Turn on the backlight since something changed on the screen
	L.internal_lcd.BacklightOn()
	L.bltimer.Reset(time.Duration(L.Backlight_secs) * time.Second)
}

func (L *LCDConfig) Clear() {
	//fmt.Println("Clearing LCD")
	L.internal_lcd.Clear()
}

func (L *LCDConfig) Close() {
	if L.internal_i2c != nil {
		L.internal_i2c.Close()
	}
}
