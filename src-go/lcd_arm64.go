//go:build arm64

package main

import (
	"fmt"
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
	hex_addr    uint8       `json:"-"`
	lcd_enabled bool        `json:"-"`
	bltimer     *time.Timer `json:"-"`
}

func (L *LCDConfig) Setup() (err error) {
	haddr, err := strconv.ParseUint(strings.TrimPrefix(L.Hex_addr, "0x"), 16, 8)
	if err != nil {
		return err
	}
	L.hex_addr = uint8(haddr)

	i_lcd, i_i2c, err := L.initLCD()
	if err != nil {
		fmt.Println("I2C LCD not configured correctly:", err)
		return err
	}
	defer i_i2c.Close()

	//Put it in "standby" mode initially
	i_lcd.Clear()
	i_lcd.BacklightOff()
	//Setup the backlight timer to turn off after a period of seconds
	L.bltimer = time.NewTimer(time.Duration(L.Backlight_secs) * time.Second)
	go func() {
		for {
			<-L.bltimer.C
			L.backlightOff()
		}
	}()
	L.bltimer.Stop() //Don't need to start initially - already off
	return err
}

func (L *LCDConfig) Display(text string) {
	if !L.lcd_enabled {
		return
	}
	ilcd, ii2c, err := L.initLCD()
	if err != nil {
		return
	}
	defer ii2c.Close()
	err = ilcd.ShowMessage(text, hd44780.SHOW_LINE_1)
	if err != nil {
		fmt.Println("Error writing to LCD: bytes written:", err)
		return
	}
	//Turn on the backlight since something changed on the screen
	ilcd.BacklightOn()
	L.bltimer.Reset(time.Duration(L.Backlight_secs) * time.Second)
}

func (L *LCDConfig) Clear() {
	if !L.lcd_enabled {
		return
	}
	ilcd, ii2c, err := L.initLCD()
	if err != nil {
		return
	}
	defer ii2c.Close()
	ilcd.Clear()
}

func (L *LCDConfig) Close() {}

func (L *LCDConfig) initLCD() (*hd44780.Lcd, *i2c.I2C, error) {
	// Create a new I2C bus connection.
	i_i2c, err := i2c.NewI2C(L.hex_addr, L.Bus_num)
	if err != nil {
		return nil, nil, err
	}

	// Create a new LCD instance.
	// We specify the I2C connection, and the LCD's dimensions (16 columns, 2 rows).
	i_lcd, err := hd44780.NewLcd(i_i2c, hd44780.LCD_16x2) // The library uses 16x2 for 1x16 displays as well.
	if err != nil {
		return nil, nil, err
	}
	return i_lcd, i_i2c, nil
}

func (L *LCDConfig) backlightOff() {
	if !L.lcd_enabled {
		return
	}
	ilcd, ii2c, err := L.initLCD()
	if err != nil {
		return
	}
	defer ii2c.Close()
	ilcd.BacklightOff()
}
