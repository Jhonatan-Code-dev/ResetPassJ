package email

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"gopkg.in/gomail.v2"
)

// =====================================================
// ⚙️ CONFIGURACIÓN GENERAL DE CORREO
// =====================================================

type EmailConfig struct {
	// ---------------- SMTP ----------------
	Host     string // Servidor SMTP (ej: smtp.gmail.com)
	Port     int    // Puerto SMTP (ej: 587 o 465)
	Username string // Correo o usuario autenticado
	Password string // Contraseña o token de aplicación

	// ---------------- DATOS DE APP ----------------
	AppName string // Nombre de la aplicación (ej: "MiApp Online")
	Title   string // Asunto del correo (ej: "Restablecimiento de Contraseña")

	// ---------------- POLÍTICAS ----------------
	CodeLength        int           // Longitud del código (ej: 6 dígitos)
	CodeValidMinutes  int           // Tiempo de validez en minutos
	MaxResetAttempts  int           // Máximo número de intentos permitidos
	RestrictionPeriod time.Duration // Tiempo de restricción tras superar los intentos
}

// =====================================================
// 📧 SERVICIO PRINCIPAL
// =====================================================

type EmailService struct {
	dialer *gomail.Dialer
	sender string
	conf   EmailConfig
}

// Instancia global (singleton)
var Service *EmailService

// Init inicializa el servicio global de correo
func Init(cfg EmailConfig) {
	Service = NewEmailService(cfg)
	log.Printf("✅ Servicio de correo '%s' inicializado correctamente. Código de %d dígitos, %d minutos de validez, %d intentos máx.",
		cfg.AppName, cfg.CodeLength, cfg.CodeValidMinutes, cfg.MaxResetAttempts)
}

// NewEmailService crea una nueva instancia del servicio
func NewEmailService(cfg EmailConfig) *EmailService {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	return &EmailService{
		dialer: dialer,
		sender: cfg.Username,
		conf:   cfg,
	}
}

// =====================================================
// 🛠️ MÉTODOS INTERNOS
// =====================================================

func (e *EmailService) send(to, subject, htmlBody string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", e.sender)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", htmlBody)

	return e.dialer.DialAndSend(msg)
}

// =====================================================
// 🔐 GENERACIÓN DE CÓDIGO DE VERIFICACIÓN
// =====================================================

// GenerateCode genera un código aleatorio de longitud configurada
func (e *EmailService) GenerateCode() string {
	digits := "0123456789"
	code := make([]byte, e.conf.CodeLength)
	for i := range code {
		code[i] = digits[rand.Intn(len(digits))]
	}
	return string(code)
}

// =====================================================
// ✉️ ENVÍO DE CORREO DE RESTABLECIMIENTO
// =====================================================

type ResetEmailData struct {
	AppName     string
	Title       string
	Code        string
	Minutes     int
	MaxAttempts int
	Restriction string
}

// SendResetPassword genera y envía el correo de restablecimiento
func (e *EmailService) SendResetPassword(to string) error {
	code := e.GenerateCode()

	data := ResetEmailData{
		AppName:     e.conf.AppName,
		Title:       e.conf.Title,
		Code:        code,
		Minutes:     e.conf.CodeValidMinutes,
		MaxAttempts: e.conf.MaxResetAttempts,
		Restriction: fmt.Sprintf("%.0f horas", e.conf.RestrictionPeriod.Hours()),
	}

	tmpl, err := template.ParseFiles("pkg/email/templates/reset_password.html")
	if err != nil {
		return fmt.Errorf("error cargando plantilla HTML: %w", err)
	}

	var htmlBody strings.Builder
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("error ejecutando plantilla HTML: %w", err)
	}

	return e.send(to, e.conf.Title, htmlBody.String())
}
