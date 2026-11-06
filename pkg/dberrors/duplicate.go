// Package dberrors classifies database errors across supported SQL dialects.
package dberrors

import (
	"errors"

	"github.com/android-sms-gateway/server/pkg/mysql"
	"gorm.io/gorm"
)

type sqlStateError interface{ SQLState() string }

// IsDuplicateKey recognizes translated GORM, MySQL and PostgreSQL errors.
func IsDuplicateKey(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) || mysql.IsDuplicateKeyViolation(err) {
		return true
	}
	var pg sqlStateError
	return errors.As(err, &pg) && pg.SQLState() == "23505"
}
