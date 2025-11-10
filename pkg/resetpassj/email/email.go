// Package email
package email

import (
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	dialer *gomail.Dialer
	sender string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

// NewEmailService permite configurar cualquier proveedor SMTP (Gmail, Outlook, Mailgun, etc.)
func NewEmailService(cfg SMTPConfig) *EmailService {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	return &EmailService{
		dialer: dialer,
		sender: cfg.Username,
	}
}

func (e *EmailService) Send(to, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", e.sender)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	return e.dialer.DialAndSend(msg)
}
