package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbFile string `json:"db_file"`
	Email *Email  `json:"email"`
}

func DefaultConfig() Config {
	return Config{
		DbFile: "test.sqlite",
		Email: &Email{
			SmtpHost: "smtp.gmail.com",
			SmtpPort: 587,
			SmtpUsername: "",
			SmtpPassword: "",
			Sender: "",
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
