package email

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

var bucketName = []byte("reset_codes")

type CodeEntry struct {
	Email    string
	Code     string
	ExpireAt time.Time
	Attempts int
	Used     bool
}

func InitDB(dbPath string) (*bbolt.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), os.ModePerm); err != nil {
		return nil, err
	}
	db, err := bbolt.Open(dbPath, 0666, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists(bucketName)
		return e
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func SaveCode(db *bbolt.DB, entry CodeEntry) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		data, _ := json.Marshal(entry)
		return b.Put([]byte(entry.Email), data)
	})
}

func GetCodeEntry(db *bbolt.DB, email string) (*CodeEntry, error) {
	var entry CodeEntry
	err := db.View(func(tx *bbolt.Tx) error {
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
