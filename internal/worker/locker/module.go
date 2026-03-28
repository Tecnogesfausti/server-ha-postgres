package locker

import (
	"context"
	"database/sql"

	"github.com/capcom6/go-infra-fx/db"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	const timeoutSeconds = 10

	return fx.Module(
		"locker",
		logger.WithNamedLogger("locker"),
		fx.Provide(func(sqlDB *sql.DB, cfg db.Config) Locker {
			switch cfg.Dialect {
			case db.DialectPostgres:
				return NewPostgresLocker(sqlDB, "worker:", timeoutSeconds)
			default:
				return NewMySQLLocker(sqlDB, "worker:", timeoutSeconds)
			}
		}),
		fx.Invoke(func(locker Locker, lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return nil
				},
				OnStop: func(_ context.Context) error {
					return locker.Close()
				},
			})
		}),
	)
}
