package migration

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/railzwaylabs/github-actions-samples/internal/platform/config"
	"go.uber.org/zap"
)

func Run(cfg config.Config, log *zap.Logger, command string) error {
	path, err := filepath.Abs(cfg.MigrationsPath)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}
	runner, err := migrate.New("file://"+filepath.ToSlash(path), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("initialize migration: %w", err)
	}
	defer func() {
		sourceErr, databaseErr := runner.Close()
		if sourceErr != nil {
			log.Warn("close migration source", zap.Error(sourceErr))
		}
		if databaseErr != nil {
			log.Warn("close migration database", zap.Error(databaseErr))
		}
	}()

	switch command {
	case "up":
		err = runner.Up()
	case "version":
		var version uint
		var dirty bool
		version, dirty, err = runner.Version()
		if err == nil {
			log.Info("database migration version", zap.Uint("version", version), zap.Bool("dirty", dirty))
		}
	default:
		return fmt.Errorf("unsupported migration command %q; use up or version", command)
	}
	if errors.Is(err, migrate.ErrNoChange) {
		log.Info("database schema already current")
		return nil
	}
	if err != nil {
		return fmt.Errorf("migrate %s: %w", command, err)
	}
	log.Info("database migration completed", zap.String("command", command))
	return nil
}
