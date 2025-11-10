package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

var (
	bucketName = []byte("reset_codes")
	dbFileName = "resetpassj.db" // Nombre fijo de la base de datos dentro del módulo
)

// CodeEntry representa un código de restablecimiento con fecha de expiración
type CodeEntry struct {
	Code     string
	ExpireAt time.Time
}

// Store encapsula la base de datos bbolt
type Store struct {
	db *bbolt.DB
}

// NewStore crea o abre la base de datos dentro de pkg/resetpassj/storage
func NewStore() (*Store, error) {
	// Obtener ruta absoluta del proyecto
	basePath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Crear carpeta storage si no existe
	storagePath := filepath.Join(basePath, "pkg", "resetpassj", "storage")
	if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
		return nil, err
	}

	// Ruta completa de la base de datos
	dbPath := filepath.Join(storagePath, dbFileName)

	db, err := bbolt.Open(dbPath, 0666, nil)
	if err != nil {
		return nil, err
	}

	// Crear bucket si no existe
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

// SaveCode guarda un código para un correo con expiración
func (s *Store) SaveCode(email, code string, expireAt time.Time) error {
	entry := CodeEntry{Code: code, ExpireAt: expireAt}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put([]byte(email), data)
	})
}

// VerifyCode verifica un código de restablecimiento
func (s *Store) VerifyCode(email, code string) (bool, error) {
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
		return false, err
	}

	if time.Now().After(entry.ExpireAt) {
		return false, errors.New("código expirado")
	}

	return entry.Code == code, nil
}
