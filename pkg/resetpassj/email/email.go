// Package email
package email

import (
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	dialer *gomail.Dialer
	user   string
}

func NewEmailService(user, pass string) *EmailService {
	dialer := gomail.NewDialer("smtp.gmail.com", 587, user, pass)
	return &EmailService{dialer: dialer, user: user}
}

func (e *EmailService) Send(to, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", e.user)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	return e.dialer.DialAndSend(msg)
}
