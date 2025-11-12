package email

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
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
	Username string // Correo o usuario autenticado (OBLIGATORIO)
	Password string // Contraseña o token de aplicación (OBLIGATORIO)

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

// Instancia global
var Service *EmailService

// =====================================================
// 🧩 CONFIGURACIÓN POR DEFECTO
// =====================================================

var defaultEmailConfig = EmailConfig{
	Host:              "smtp.gmail.com",
	Port:              587,
	AppName:           "MiApp",
	Title:             "Restablecimiento de contraseña",
	CodeLength:        6,
	CodeValidMinutes:  15,
	MaxResetAttempts:  3,
	RestrictionPeriod: 24 * time.Hour,
}

// =====================================================
// 🚀 INICIALIZACIÓN
// =====================================================

func Init(cfg EmailConfig) error {
	if cfg.Username == "" || cfg.Password == "" {
		return errors.New("❌ 'Username' y 'Password' son obligatorios para inicializar el servicio de correo")
	}

	// Aplicar valores por defecto
	if cfg.Host == "" {
		cfg.Host = defaultEmailConfig.Host
	}
	if cfg.Port == 0 {
		cfg.Port = defaultEmailConfig.Port
	}
	if cfg.AppName == "" {
		cfg.AppName = defaultEmailConfig.AppName
	}
	if cfg.Title == "" {
		cfg.Title = defaultEmailConfig.Title
	}
	if cfg.CodeLength == 0 {
		cfg.CodeLength = defaultEmailConfig.CodeLength
	}
	if cfg.CodeValidMinutes == 0 {
		cfg.CodeValidMinutes = defaultEmailConfig.CodeValidMinutes
	}
	if cfg.MaxResetAttempts == 0 {
		cfg.MaxResetAttempts = defaultEmailConfig.MaxResetAttempts
	}
	if cfg.RestrictionPeriod == 0 {
		cfg.RestrictionPeriod = defaultEmailConfig.RestrictionPeriod
	}

	Service = NewEmailService(cfg)

	log.Printf("✅ Servicio de correo '%s' inicializado correctamente.\n", cfg.AppName)
	log.Printf("   ➤ Código: %d dígitos | Validez: %d min | Intentos: %d | Restricción: %.0f horas",
		cfg.CodeLength, cfg.CodeValidMinutes, cfg.MaxResetAttempts, cfg.RestrictionPeriod.Hours())

	return nil
}

// =====================================================
// 🏗️ CONSTRUCTOR
// =====================================================

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

func (e *EmailService) GenerateCode() string {
	rand.Seed(time.Now().UnixNano())
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

	// ✅ Construcción dinámica de la ruta absoluta del HTML (versión corregida)
	execDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error obteniendo directorio actual: %w", err)
	}

	// Ajusta la ruta según tu estructura real:
	templatePath := filepath.Join(execDir, "pkg", "resetpassj", "email", "templates", "reset_password.html")

	// Verifica existencia antes de abrir
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("la plantilla no existe en: %s", templatePath)
	}

	// Cargar la plantilla
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("error cargando plantilla HTML desde %s: %w", templatePath, err)
	}

	var htmlBody strings.Builder
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("error ejecutando plantilla HTML: %w", err)
	}

	return e.send(to, e.conf.Title, htmlBody.String())
}
