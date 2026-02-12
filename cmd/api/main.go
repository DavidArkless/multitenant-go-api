package main

import (
	"context"
	"log/slog"
	"multitenant-go-api/internal/api"
	"multitenant-go-api/internal/apps/domain/authapp"
	"multitenant-go-api/internal/apps/domain/healthapp"
	"multitenant-go-api/internal/apps/domain/tenantapp"
	"multitenant-go-api/internal/apps/domain/userapp"
	"multitenant-go-api/internal/auth"
	"multitenant-go-api/internal/config"
	"multitenant-go-api/internal/seed"
	"os"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	err := run(context.Background(), logger)
	if err != nil {
		panic(err)
	}
}

func run(ctx context.Context, log *slog.Logger) error {
	log.Info("Starting Service ")

	// Load Config
	cfg := config.NewConfig()

	log.Info("Config loaded", "config", cfg)

	// Start DB
	log.Info("initializing database support", "hostport", cfg.DB.Host)

	db, err := sqlx.Connect("sqlite", cfg.DB.Host)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := seed.SeedIfEmpty(ctx, db, seed.Options{Force: false}); err != nil {
		return err
	}

	// Setup Debug

	// Setup Repository
	tenantRepo := tenantapp.NewSQLiteProjectRepo(db)
	userRepo := userapp.NewSQLiteUserRepo(db)

	// Setup services
	jwtService := auth.NewJWTService(cfg.Auth.Key)
	health := healthapp.NewHealthApp(log)
	tenantApp := tenantapp.New(log, tenantRepo)
	authApp := authapp.New(log, jwtService, userRepo)

	// Setup API

	apps := api.Apps{
		HealthApp: health,
		TenantApp: tenantApp,
		AuthApp:   authApp,
	}

	authentication := api.Authentication{
		JwtService: jwtService,
	}

	server := api.NewApi(cfg.App, log, apps, authentication)

	return server.Start()

}
