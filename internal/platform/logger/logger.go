package logger

import (
	"context"

	"github.com/railzwaylabs/github-actions-samples/internal/platform/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("platform.logger", fx.Provide(New))

func New(lifecycle fx.Lifecycle, cfg config.Config) (*zap.Logger, error) {
	var log *zap.Logger
	var err error
	if cfg.Environment == "development" {
		log, err = zap.NewDevelopment()
	} else {
		log, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		_ = log.Sync()
		return nil
	}})
	return log, nil
}
