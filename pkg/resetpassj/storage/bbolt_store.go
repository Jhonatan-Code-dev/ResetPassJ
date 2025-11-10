package storage

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

var (
	bucketName = []byte("reset_codes")
	dbFileName = "resetpassj.db"
)

// CodeEntry representa un código de restablecimiento por correo
type CodeEntry struct {
	Email       string
	Code        string
	ExpireAt    time.Time
	Attempts    int
	MaxAttempts int
	Used        bool
}

// Store encapsula la base de datos bbolt
type Store struct {
	db *bbolt.DB
}

// NewStore crea o abre la base de datos dentro de pkg/resetpassj/storage
func NewStore() (*Store, error) {
	basePath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	storagePath := filepath.Join(basePath, "pkg", "resetpassj", "storage")
	if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(storagePath, dbFileName)

	db, err := bbolt.Open(dbPath, 0666, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

// Close cierra la base de datos
func (s *Store) Close() {
	s.db.Close()
}

// generateCode genera un código aleatorio de longitud n
func generateCode(length int) string {
	const charset = "0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (s *Store) SaveCode(email string, codeLength int, expireMinutes int, maxAttempts int) (string, error) {
	// Valores por defecto
	if codeLength <= 0 {
		codeLength = 4
	}
	if expireMinutes <= 0 {
		expireMinutes = 2
	}
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	code := generateCode(codeLength)

	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)

		var entry CodeEntry
		data := b.Get([]byte(email))
		if data != nil {
			if err := json.Unmarshal(data, &entry); err != nil {
				return err
			}
			// Sumar un intento por generar nuevo código, hasta MaxAttempts
			entry.Attempts++
			if entry.Attempts > maxAttempts {
				entry.Attempts = maxAttempts
			}
		} else {
			// Si no existía, inicializar Attempts en 1
			entry.Attempts = 1
		}

		// Actualizar registro con nuevo código y expiración
		entry.Email = email
		entry.Code = code
		entry.ExpireAt = time.Now().Add(time.Duration(expireMinutes) * time.Minute)
		entry.Used = false
		entry.MaxAttempts = maxAttempts

		newData, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		return b.Put([]byte(email), newData)
	})

	if err != nil {
		return "", err
	}
	return code, nil
}

// VerifyCode verifica un código y aumenta intentos, marca como usado si es correcto
func (s *Store) VerifyCode(email, code string) (bool, error) {
	var entry CodeEntry
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		data := b.Get([]byte(email))
		if data == nil {
			return errors.New("correo no encontrado")
		}

		if err := json.Unmarshal(data, &entry); err != nil {
			return err
		}

		if entry.Used {
			return errors.New("código ya usado")
		}

		// Incrementar intentos hasta MaxAttempts
		if entry.Attempts < entry.MaxAttempts {
			entry.Attempts++
		}

		// Validar código
		if entry.Code == code {
			entry.Used = true
		}

		updatedData, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		return b.Put([]byte(email), updatedData)
	})
	if err != nil {
		return false, err
	}

	if time.Now().After(entry.ExpireAt) {
		return false, errors.New("código expirado")
	}

	return entry.Code == code, nil
}

// GetCodeEntry devuelve el registro completo de un correo
func (s *Store) GetCodeEntry(email string) (*CodeEntry, error) {
	var entry CodeEntry
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		data := b.Get([]byte(email))
		if data == nil {
			return errors.New("correo no encontrado")
		}
		return json.Unmarshal(data, &entry)
	})
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// DumpDB devuelve todos los registros del bucket tal como están guardados en la base de datos
func (s *Store) DumpDB() (map[string][]byte, error) {
	records := make(map[string][]byte)

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return errors.New("bucket no encontrado")
		}

		return b.ForEach(func(k, v []byte) error {
			// Guardamos cada registro tal como está
			records[string(k)] = append([]byte(nil), v...)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return records, nil
}
