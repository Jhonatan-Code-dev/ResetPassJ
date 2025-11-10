package logic

import (
	"ResetPassJ/pkg/resetpassj/email"
	"ResetPassJ/pkg/resetpassj/storage"
	"crypto/rand"
	"fmt"
)

type ResetManager struct {
	store storage.Store
	email *email.EmailService
}

func generateRandomCode() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "000000" // fallback
	}
	return fmt.Sprintf("%06d", int(b[0])%1000000)
}
