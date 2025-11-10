package logic

import (
	"ResetPassJ/pkg/resetpassj/email"
	"ResetPassJ/pkg/resetpassj/storage"
	"crypto/rand"
	"fmt"
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

func generateRandomCode() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "000000" // fallback
	}
	return fmt.Sprintf("%06d", int(b[0])%1000000)
}
