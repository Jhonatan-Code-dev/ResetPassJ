package logic

import (
	"ResetPassJ/pkg/resetpassj/email"
	"time"
)

type ResetManager struct {
	store storage.Store
	email *email.EmailService
}

// Genera un código, lo guarda en bbolt y lo envía por correo
func (rm *ResetManager) SendResetCode(to string) error {
	code := generateRandomCode()
	rm.store.SaveCode(to, code, time.Now().Add(2*time.Minute))
	body := "Tu código de restablecimiento es: " + code
	return rm.email.Send(to, "Restablecimiento de contraseña", body)
}
