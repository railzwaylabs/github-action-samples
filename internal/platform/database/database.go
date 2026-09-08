package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/railzwaylabs/github-actions-samples/internal/platform/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Module = fx.Module("platform.database", fx.Provide(New))

func New(lifecycle fx.Lifecycle, cfg config.Config, log *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{Logger: newGORMLogger(log)})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("access database pool: %w", err)
	}
	configurePool(sqlDB)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { return sqlDB.Close() }})
	return db, nil
}

func configurePool(db *sql.DB) {
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
}

func newGORMLogger(log *zap.Logger) logger.Interface {
	return logger.New(zapWriter{log: log}, logger.Config{SlowThreshold: 500 * time.Millisecond, LogLevel: logger.Warn, Colorful: false})
}

type zapWriter struct{ log *zap.Logger }

func (w zapWriter) Printf(message string, values ...any) {
	w.log.Warn("gorm", zap.String("message", fmt.Sprintf(message, values...)))
}
