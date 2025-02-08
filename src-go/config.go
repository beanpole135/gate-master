package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Host     string     `json:"host_url"`
	SiteName string     `json:"site_name"`
	DbFile   string     `json:"db_file"`
	Auth     AuthConfig `json:"auth"`
	Email    *Email     `json:"email"`
	Keypad   *Keypad    `json:"keypad_pins"`
}
type AuthConfig struct {
	HashKey      string `json:"hash_key"`
	BlockKey     string `json:"block_key"`
	JwtSecret    string `json:"jwtsecret"`
	JwtTokenSecs int    `json:"jwttokensecs"` //lifetime in seconds
}

func DefaultConfig() Config {
	return Config{
		Host:     "http://localhost:8080",
		SiteName: "Shadow Mountain",
		DbFile:   "test.sqlite",
		Auth: AuthConfig{
			JwtSecret:    "testkey",
			JwtTokenSecs: 3600,
		},
		Email: &Email{
			SmtpHost:     "smtp.gmail.com",
			SmtpPort:     587,
			SmtpUsername: "",
			SmtpPassword: "",
			Sender:       "",
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
	return &C, err
}
