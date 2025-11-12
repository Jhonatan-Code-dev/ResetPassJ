package logic

import (
	"crypto/rand"
	"fmt"

	"github.com/Jhonatan-Code-dev/ResetPassJ/pkg/resetpassj/email"
	"github.com/Jhonatan-Code-dev/ResetPassJ/pkg/resetpassj/storage"
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
