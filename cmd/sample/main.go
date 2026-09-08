package main

import (
	"fmt"
	"os"

	healthhttp "github.com/railzwaylabs/github-actions-samples/internal/health/transport/http"
	"github.com/railzwaylabs/github-actions-samples/internal/order/application"
	"github.com/railzwaylabs/github-actions-samples/internal/order/infrastructure/persistent"
	orderhttp "github.com/railzwaylabs/github-actions-samples/internal/order/transport/http"
	"github.com/railzwaylabs/github-actions-samples/internal/platform/config"
	"github.com/railzwaylabs/github-actions-samples/internal/platform/database"
	platformhttp "github.com/railzwaylabs/github-actions-samples/internal/platform/http"
	"github.com/railzwaylabs/github-actions-samples/internal/platform/logger"
	"github.com/railzwaylabs/github-actions-samples/internal/platform/migration"
	"github.com/railzwaylabs/github-actions-samples/internal/platform/profiling"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigration()
		return
	}
	fx.New(
		fx.Provide(func() (config.Config, error) { return config.New(version) }),
		logger.Module,
		database.Module,
		platformhttp.Module,
		profiling.Module,
		persistent.Module,
		application.Module,
		orderhttp.Module,
		healthhttp.Module,
	).Run()
}

func runMigration() {
	command := "up"
	if len(os.Args) > 2 {
		command = os.Args[2]
	}
	cfg, err := config.New(version)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	log, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() { _ = log.Sync() }()
	if err := migration.Run(cfg, log, command); err != nil {
		log.Error("migration failed", zap.Error(err))
		os.Exit(1)
	}
}
