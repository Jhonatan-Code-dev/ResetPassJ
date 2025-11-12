package email

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
	"gopkg.in/gomail.v2"
)

type EmailConfig struct {
	Host              string
	Port              int
	Username          string
	Password          string
	AppName           string
	Title             string
	CodeLength        int
	CodeValidMinutes  int
	MaxResetAttempts  int
	RestrictionPeriod time.Duration
}

type EmailService struct {
	dialer *gomail.Dialer
	sender string
	conf   EmailConfig
	db     *bbolt.DB
}

var Service *EmailService

func Init(cfg EmailConfig) error {
	if cfg.Username == "" || cfg.Password == "" {
		return errors.New("❌ 'Username' y 'Password' son obligatorios")
	}
	applyDefaults(&cfg)

	baseDir, _ := os.Getwd()
	dbPath := filepath.Join(baseDir, "pkg", "resetpassj", "storage", "resetpassj.db")

	db, err := InitDB(dbPath)
	if err != nil {
		return fmt.Errorf("error iniciando base de datos: %w", err)
	}

	Service = &EmailService{
		dialer: gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password),
		sender: cfg.Username,
		conf:   cfg,
		db:     db,
	}

	log.Printf("✅ Servicio '%s' listo | Código: %d dígitos | Validez: %d min | Intentos: %d",
		cfg.AppName, cfg.CodeLength, cfg.CodeValidMinutes, cfg.MaxResetAttempts)
	return nil
}

func applyDefaults(cfg *EmailConfig) {
	def := EmailConfig{
		Host:              "smtp.gmail.com",
		Port:              587,
		AppName:           "MiApp",
		Title:             "Restablecimiento de contraseña",
		CodeLength:        6,
		CodeValidMinutes:  15,
		MaxResetAttempts:  3,
		RestrictionPeriod: 24 * time.Hour,
	}
	if cfg.Host == "" {
		cfg.Host = def.Host
	}
	if cfg.Port == 0 {
		cfg.Port = def.Port
	}
	if cfg.AppName == "" {
		cfg.AppName = def.AppName
	}
	if cfg.Title == "" {
		cfg.Title = def.Title
	}
	if cfg.CodeLength == 0 {
		cfg.CodeLength = def.CodeLength
	}
	if cfg.CodeValidMinutes == 0 {
		cfg.CodeValidMinutes = def.CodeValidMinutes
	}
	if cfg.MaxResetAttempts == 0 {
		cfg.MaxResetAttempts = def.MaxResetAttempts
	}
	if cfg.RestrictionPeriod == 0 {
		cfg.RestrictionPeriod = def.RestrictionPeriod
	}
}
