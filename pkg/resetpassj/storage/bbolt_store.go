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

type CodeEntry struct {
	Email    string
	Code     string
	ExpireAt time.Time
	Attempts int // Se incrementa solo en SaveCode
	Used     bool
}

type Store struct {
	db *bbolt.DB
}

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

func (s *Store) Close() {
	s.db.Close()
}

// --- Genera código aleatorio ---
func generateCode(length int) string {
	const digits = "0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

// --- SaveCode con LÍMITE de intentos ---
func (s *Store) SaveCode(email string, codeLength int, expireMinutes int, maxAttempts int) (string, error) {
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
			json.Unmarshal(data, &entry)

			// 🚫 No permitir más generación si llegó al límite
			if entry.Attempts >= maxAttempts {
				return errors.New("has alcanzado el límite de solicitudes de código")
			}

			entry.Attempts++
		} else {
			entry.Attempts = 1
		}

		entry.Email = email
		entry.Code = code
		entry.ExpireAt = time.Now().Add(time.Duration(expireMinutes) * time.Minute)
		entry.Used = false

		newData, _ := json.Marshal(entry)
		return b.Put([]byte(email), newData)
	})

	if err != nil {
		return "", err
	}

	return code, nil
}

// --- VerifyCode NO modifica Attempts ---
func (s *Store) VerifyCode(email, code string) (bool, error) {
	var entry CodeEntry

	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		data := b.Get([]byte(email))
		if data == nil {
			return errors.New("correo no encontrado")
		}
		json.Unmarshal(data, &entry)

		if time.Now().After(entry.ExpireAt) {
			return errors.New("código expirado")
		}
		if entry.Used {
			return errors.New("código ya usado")
		}

		if entry.Code == code {
			entry.Used = true
		}

		newData, _ := json.Marshal(entry)
		return b.Put([]byte(email), newData)
	})

	if err != nil {
		return false, err
	}

	return entry.Code == code, nil
}

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

func (s *Store) DumpDB() (map[string][]byte, error) {
	records := make(map[string][]byte)

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.ForEach(func(k, v []byte) error {
			records[string(k)] = append([]byte(nil), v...)
			return nil
		})
	})
	return records, err
}
