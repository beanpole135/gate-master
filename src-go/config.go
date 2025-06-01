package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	filepath string     `json:"-"` //internal for where the file was loaded from
	Host     string     `json:"host_url"`
	SiteName string     `json:"site_name"`
	DbFile   string     `json:"db_file"`
	LogsDir  string     `json:"logs_directory"`
	Auth     AuthConfig `json:"auth"`
	Email    *Email     `json:"email"`
	Keypad   *Keypad    `json:"keypad_pins"`
	Camera   CamConfig  `json:"camera"`
	Gate     GateConfig `json:"gate"`
	LCD      LCDConfig  `json:"lcd_i2c"`
}
type AuthConfig struct {
	HashKey      string `json:"hash_key"`
	BlockKey     string `json:"block_key"`
	JwtSecret    string `json:"jwtsecret"`
	JwtTokenSecs int    `json:"jwttokensecs"` //lifetime in seconds
}

func DefaultConfig() Config {
	return Config{
		Host:     ":8080",
		SiteName: "Gate Control",
		DbFile:   "test.sqlite",
		LogsDir:  "",
		Auth: AuthConfig{
			JwtSecret:    "",
			JwtTokenSecs: 3600,
		},
		Email: &Email{
			SmtpHost:     "smtp.gmail.com",
			SmtpPort:     587,
			SmtpUsername: "",
			SmtpPassword: "",
			Sender:       "",
		},
		Camera: CamConfig{
			Rotation: 0,
			Width:    1024,
			Height:   768,
		},
		LCD: LCDConfig{
			Bus_num:        1,
			Backlight_secs: 10,
			Hex_addr:       "0x27",
			Pins: I2CPins{
				En:        6,
				Rw:        5,
				Rs:        4,
				D4:        11,
				D5:        12,
				D6:        13,
				D7:        14,
				Backlight: 3,
			},
		},
	}
}
func LoadConfig(path string) (*Config, error) {
	C := DefaultConfig()
	body, err := os.ReadFile(path)
	if err != nil {
		return &C, err
	}
	err = json.Unmarshal(body, &C)
	C.filepath = path //save internally for later
	return &C, err
}

func UpdateConfig(C *Config) {
	if C == nil || C.filepath == "" {
		return
	}
	body, err := json.MarshalIndent(C, "", "  ")
	if err == nil {
		err = os.WriteFile(C.filepath, body, 0600)
	}
	if err != nil {
		fmt.Println("Unable to update config file: ", C.filepath)
	}
}
