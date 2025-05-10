//go:build arm64

package main

import (
	"fmt"
	"time"

	"github.com/davecheney/i2c"
)

type LCDConfig struct {
	//Config file variables
	backlight_secs int `json:"backlight_seconds"`

	//Internal variables
	internal_i2c *i2c.I2C    `json:"-"`
	internal_lcd *i2c.Lcd    `json:"-"`
	bltimer      *time.Timer `json:"-"`
}

func (L *LCDConfig) Setup() (err error) {
	L.internal_i2c, err = i2c.New(0x27, 1)
	if err != nil {
		return err
	}
	//Config Pins input order: [en, rw, rs, d4, d5, d6, d7, backlight]
	L.internal_lcd, err = i2c.NewLcd(L.internal_i2c, 2, 1, 0, 4, 5, 6, 7, 3)
	//Put it in "standby" mode initially
	L.internal_lcd.Clear()
	L.internal_lcd.BacklightOff()
	//Setup the backlight timer to turn off after a period of seconds
	L.bltimer = time.NewTimer(time.Duration(L.backlight_secs) * time.Second)
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
	//Put the text on the screen
	L.internal_lcd.Clear()
	L.internal_lcd.SetPosition(1, 0) //line 1, character 0
	fmt.Fprintf(L.internal_lcd, text)

	//Turn on the backlight since something changed on the screen
	L.internal_lcd.BacklightOn()
	L.bltimer.Reset(time.Duration(L.backlight_secs) * time.Second)
}

func (L *LCDConfig) Clear() {
	L.internal_lcd.Clear()
}
