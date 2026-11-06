package incoming

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type errorDriver struct{ err error }

func (d errorDriver) Open(string) (driver.Conn, error) { return errorConn{d.err}, nil }

type errorConn struct{ err error }

func (c errorConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unexpected prepare") }
func (c errorConn) Close() error                        { return nil }
func (c errorConn) Begin() (driver.Tx, error)           { return nil, errors.New("unexpected transaction") }
func (c errorConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, c.err
}
func (c errorConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return nil, c.err
}

func TestIncomingRetryHandlesPostgresDuplicate(t *testing.T) {
	for _, tc := range []struct {
		name, code string
		wantError  bool
	}{
		{"unique violation is idempotent", "23505", false},
		{"foreign key violation is an error", "23503", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// A dedicated database/sql connector simulates the actual pgx error; no production DB is touched.
			db := sql.OpenDB(errorConnector{errorDriver{&pgconn.PgError{Code: tc.code}}})
			defer db.Close()
			gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{
				DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.New(logWriter{t}, logger.Config{LogLevel: logger.Silent}),
			})
			if err != nil {
				t.Fatal(err)
			}
			service := NewService(NewRepository(gdb))
			err = service.Insert(CreateMessageInput{ID: "stable-id", DeviceID: "device", Type: MessageTypeSMS, Sender: "sender", ContentPreview: "text", ReceivedAt: time.Now()})
			if (err != nil) != tc.wantError {
				t.Fatalf("Insert returned error %v; want error %v", err, tc.wantError)
			}
		})
	}
}

type errorConnector struct{ d errorDriver }

func (c errorConnector) Connect(context.Context) (driver.Conn, error) { return c.d.Open("") }
func (c errorConnector) Driver() driver.Driver                        { return c.d }

type logWriter struct{ t *testing.T }

func (l logWriter) Printf(string, ...interface{}) {}
