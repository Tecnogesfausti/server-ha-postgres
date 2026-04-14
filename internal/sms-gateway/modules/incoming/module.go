package incoming

import (
	"github.com/capcom6/go-infra-fx/db"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"incoming",
		logger.WithNamedLogger("incoming"),
		fx.Provide(
			NewRepository,
			fx.Private,
		),
		fx.Provide(NewService),
	)
}

//nolint:gochecknoinits //backward compatibility
func init() {
	db.RegisterMigration(Migrate)
}
