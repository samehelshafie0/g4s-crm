package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"g4s-crm/api/internal/bootstrap"
	"g4s-crm/api/internal/config"
	"g4s-crm/api/internal/database"
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/router"
	"g4s-crm/api/migrations"
	pkgvalidator "g4s-crm/api/pkg/validator"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// ─── Logging ─────────────────────────────────────────────
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("APP_ENV") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	// ─── Config ──────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	level, _ := zerolog.ParseLevel(cfg.Log.Level)
	zerolog.SetGlobalLevel(level)

	log.Info().
		Str("env", cfg.App.Env).
		Str("port", cfg.App.Port).
		Msg("Starting G4S CRM API")

	// ─── Database ────────────────────────────────────────────
	db, err := database.Connect(&cfg.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	log.Info().Str("host", cfg.DB.Host).Str("name", cfg.DB.Name).Msg("Database connected")

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			err = migrations.Up(context.Background(), sqlDB)
		case "bootstrap-admin":
			err = bootstrap.Admin(db, os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD"), os.Getenv("ADMIN_FIRST_NAME"), os.Getenv("ADMIN_LAST_NAME"))
		default:
			log.Fatal().Msg("Unknown command; use migrate or bootstrap-admin")
		}
		if err != nil {
			log.Fatal().Err(err).Msg("Command failed")
		}
		log.Info().Msg("Command completed")
		return
	}
	if err := migrations.Ready(context.Background(), sqlDB); err != nil {
		log.Fatal().Err(err).Msg("Database migrations are not ready; run migrate first")
	}

	// ─── Load JWT public key ──────────────────────────────────
	if err := middleware.LoadPublicKey(cfg.JWT.PublicKeyPath); err != nil {
		log.Fatal().Err(err).Str("path", cfg.JWT.PublicKeyPath).Msg("Failed to load JWT public key")
	}

	// ─── Validator ───────────────────────────────────────────
	pkgvalidator.Setup()

	// ─── Storage directory ───────────────────────────────────
	if err := os.MkdirAll(cfg.Storage.Root, 0750); err != nil {
		log.Fatal().Err(err).Str("path", cfg.Storage.Root).Msg("Failed to create storage directory")
	}

	// ─── Router ──────────────────────────────────────────────
	r := router.Setup(db, cfg)

	// ─── HTTP Server ─────────────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ─── Graceful Shutdown ───────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("Server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced shutdown")
	}

	log.Info().Msg("Server exited")
}
