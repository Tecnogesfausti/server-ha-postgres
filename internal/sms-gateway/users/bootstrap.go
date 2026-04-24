package users

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func registerBootstrapUser(
	lc fx.Lifecycle,
	cfg Config,
	svc *Service,
	log *zap.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			created, err := svc.EnsureExists(cfg.InternalUsername, cfg.InternalPassword)
			if err != nil {
				return fmt.Errorf("failed to ensure internal user: %w", err)
			}
			if created {
				log.Info("internal bootstrap user created", zap.String("username", cfg.InternalUsername))
			}
			return nil
		},
	})
}
