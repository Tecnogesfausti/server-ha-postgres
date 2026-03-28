package locker

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

type postgresLocker struct {
	db *sql.DB

	prefix  string
	timeout time.Duration

	mu    sync.Mutex
	conns map[string]*sql.Conn
}

func NewPostgresLocker(db *sql.DB, prefix string, timeoutSeconds int) Locker {
	return &postgresLocker{
		db:      db,
		prefix:  prefix,
		timeout: time.Duration(timeoutSeconds) * time.Second,
		mu:      sync.Mutex{},
		conns:   make(map[string]*sql.Conn),
	}
}

func (p *postgresLocker) AcquireLock(ctx context.Context, key string) error {
	name := p.prefix + key

	conn, err := p.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get conn: %w", err)
	}

	deadline := time.Now().Add(p.timeout)
	for {
		var acquired bool
		lockErr := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(hashtext($1))", name).Scan(&acquired)
		if lockErr != nil {
			_ = conn.Close()
			return fmt.Errorf("failed to get lock: %w", lockErr)
		}
		if acquired {
			p.mu.Lock()
			if prev, ok := p.conns[key]; ok && prev != nil {
				_ = prev.Close()
			}
			p.conns[key] = conn
			p.mu.Unlock()
			return nil
		}
		if time.Now().After(deadline) {
			_ = conn.Close()
			return ErrLockNotAcquired
		}

		select {
		case <-ctx.Done():
			_ = conn.Close()
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func (p *postgresLocker) ReleaseLock(ctx context.Context, key string) error {
	name := p.prefix + key

	p.mu.Lock()
	conn := p.conns[key]
	delete(p.conns, key)
	p.mu.Unlock()
	if conn == nil {
		return fmt.Errorf("%w: no held connection for key %q", ErrLockNotAcquired, key)
	}

	var unlocked bool
	err := conn.QueryRowContext(ctx, "SELECT pg_advisory_unlock(hashtext($1))", name).Scan(&unlocked)
	_ = conn.Close()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	if !unlocked {
		return fmt.Errorf("%w: lock was not held or doesn't exist", ErrLockNotAcquired)
	}

	return nil
}

func (p *postgresLocker) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for key, conn := range p.conns {
		if conn != nil {
			_ = conn.Close()
		}
		delete(p.conns, key)
	}
	return nil
}

var _ Locker = (*postgresLocker)(nil)
