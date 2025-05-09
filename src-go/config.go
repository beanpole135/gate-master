package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	filepath string      `json:"-"` //internal for where the file was loaded from
	Host     string      `json:"host_url"`
	SiteName string      `json:"site_name"`
	DbFile   string      `json:"db_file"`
	Auth     AuthConfig  `json:"auth"`
	Email    *Email      `json:"email"`
	Keypad   *Keypad     `json:"keypad_pins"`
	Camera   CamConfig   `json:"camera"`
	Gate     *GateConfig `json:"gate"`
}
type AuthConfig struct {
	HashKey      string `json:"hash_key"`
	BlockKey     string `json:"block_key"`
	JwtSecret    string `json:"jwtsecret"`
	JwtTokenSecs int    `json:"jwttokensecs"` //lifetime in seconds
}

func DefaultConfig() Config {
	return Config{
		Host:     "http://localhost:8081",
		SiteName: "Gate Control",
		DbFile:   "test.sqlite",
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
			Device:      "/dev/video0",
			PixelFormat: "mjpeg",
			Width:       1024,
			Height:      768,
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
