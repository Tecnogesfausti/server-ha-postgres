package dberrors

import (
	"errors"
	"fmt"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestDuplicateKeys(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"postgres wrapped unique", fmt.Errorf("insert: %w", &pgconn.PgError{Code: "23505"}), true},
		{"postgres other", &pgconn.PgError{Code: "23503"}, false},
		{"mysql unique", &mysqlDriver.MySQLError{Number: 1062}, true},
		{"gorm translated", gorm.ErrDuplicatedKey, true},
		{"unrelated", errors.New("failure"), false},
		{"nil", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsDuplicateKey(tc.err); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
