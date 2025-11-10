package storage

import "time"

type Store interface {
	SaveCode(email, code string, expiration time.Time) error
}
