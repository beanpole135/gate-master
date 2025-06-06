package main

import (
	"log"

	"github.com/d2r2/go-hd44780"
	"github.com/d2r2/go-i2c"
)

func main() {
	// Create a new I2C bus connection.
	// The '1' corresponds to /dev/i2c-1, which is the default on a Raspberry Pi 4B.
	i2c, err := i2c.NewI2C(0x27, 1) // Replace 0x27 with your LCD's address if different.
	if err != nil {
		log.Fatal(err)
	}
	defer i2c.Close()

	// Create a new LCD instance.
	// We specify the I2C connection, and the LCD's dimensions (16 columns, 1 row).
	lcd, err := hd44780.NewLcd(i2c, hd44780.LCD_16x2) // The library uses 16x2 for 1x16 displays as well.
	if err != nil {
		log.Fatal(err)
	}

	// Turn on the backlight.
	err = lcd.BacklightOn()
	if err != nil {
		log.Fatal(err)
	}

	// Clear the display.
	err = lcd.Clear()
	if err != nil {
		log.Fatal(err)
	}

	// Display the text on the first line.
	err = lcd.ShowMessage("Hello, World!", hd44780.SHOW_LINE_1)
	if err != nil {
		log.Fatal(err)
	}
}
