package email

import (
	"log"
	"strings"
	"text/template"

	"gopkg.in/gomail.v2"
)

// ------------------ CONFIGURACIONES ------------------

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type AppEmailConfig struct {
	AppName string
	Title   string
	Minutes int
}

// ------------------ SERVICIO PRINCIPAL ------------------

type EmailService struct {
	dialer  *gomail.Dialer
	sender  string
	appConf AppEmailConfig
}

// Instancia global (singleton)
var Service *EmailService

// Init inicializa el servicio global de correo
func Init(smtpCfg SMTPConfig, appConf AppEmailConfig) {
	Service = NewEmailService(smtpCfg, appConf)
	log.Println("✅ Servicio de correo inicializado correctamente.")
}

func NewEmailService(cfg SMTPConfig, appConf AppEmailConfig) *EmailService {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	return &EmailService{
		dialer:  dialer,
		sender:  cfg.Username,
		appConf: appConf,
	}
}

// ------------------ MÉTODO GENERAL ------------------

func (e *EmailService) send(to, subject, htmlBody string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", e.sender)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", htmlBody)

	return e.dialer.DialAndSend(msg)
}

// ------------------ ENVÍO DE CORREO RESET ------------------

type ResetEmailData struct {
	AppName string
	Title   string
	Code    string
	Minutes int
}

// SendResetPassword genera el HTML y envía el correo de recuperación
func (e *EmailService) SendResetPassword(to string, code string) error {
	data := ResetEmailData{
		AppName: e.appConf.AppName,
		Title:   e.appConf.Title,
		Code:    code,
		Minutes: e.appConf.Minutes,
	}

	tmpl, err := template.ParseFiles("pkg/email/templates/reset_password.html")
	if err != nil {
		return err
	}

	var htmlBody strings.Builder
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		return err
	}

	subject := e.appConf.Title
	return e.send(to, subject, htmlBody.String())
}
