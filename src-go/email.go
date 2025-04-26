package main

import (
	"fmt"
	"net/mail"

	gomail "gopkg.in/mail.v2"
)

func isValidEmail(eml string) bool {
	//The mail package has a parser that is RFC 5322 compliant built-in
	addr, err := mail.ParseAddress(eml)
	return (err == nil && addr.Address == eml)
}

// Gmail Reference: https://support.google.com/a/answer/176600?hl=en
type Email struct {
	SmtpHost     string `json:"smtp_host"`    // smtp.gmail.com
	SmtpPort     int    `json:"smtp_port"`    // 587 for gmail
	SmtpUsername string `json:"smtp_user"`    // Gmail address (me@gmail.com)
	SmtpPassword string `json:"smtp_pass"`    // Password for user (abcd1234)
	Sender       string `json:"sender_email"` //Typically the same as SmtpUsername
}

func (E *Email) SendEmail(to string, subject string, body string, initialEmail bool) error {
	if E.SmtpHost == "" || E.SmtpPort == 0 || E.SmtpUsername == "" {
		fmt.Println("Email system not configured")
		return nil //do nothing - email system not setup
	}
	if !initialEmail && len(body) > 150 {
		//Body of message too long
		return fmt.Errorf("Body of email is too long. 150 character limit")
	}
	// Create a new message
	message := gomail.NewMessage()
	// Set email headers
	message.SetHeader("From", E.Sender)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	// Set email body
	message.SetBody("text/plain", body)

	// Set up the SMTP dialer
	dialer := gomail.NewDialer(E.SmtpHost, E.SmtpPort, E.SmtpUsername, E.SmtpPassword)
	dialer.StartTLSPolicy = gomail.MandatoryStartTLS //Require TLS encryption for transit
	// Send the email
	if err := dialer.DialAndSend(message); err != nil {
		fmt.Println("Email submission error:", err)
		return err
	}
	return nil
}
