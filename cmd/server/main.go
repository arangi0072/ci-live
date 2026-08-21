package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ci-live/internal/auth"
	"ci-live/internal/config"
	"ci-live/internal/database"
	"ci-live/internal/router"
	"ci-live/migrations"
)

func main() {
	// --------------------------------------------------
	// 1. Load configuration
	// --------------------------------------------------

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// --------------------------------------------------
	// 2. Connect to PostgreSQL
	// --------------------------------------------------

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer db.Close()

	log.Println("connected to PostgreSQL")

	// --------------------------------------------------
	// Run Migrations
	// --------------------------------------------------

	if _, err := db.Exec(migrations.Schema); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}
	log.Println("database migrations applied successfully")

	// --------------------------------------------------
	// 3. Initialize Auth layers
	// --------------------------------------------------

	authRepository := auth.NewRepository(db)
	jwtManager := auth.NewJWTManager(
		cfg.JWTSecret,
		cfg.AccessTokenExpiry,
		cfg.RefreshTokenExpiry,
		"ci-live",
	)
	authService := auth.NewService(authRepository, jwtManager)
	authHandler := auth.NewHandler(authService)

	// --------------------------------------------------
	// 4. Initialize router
	// --------------------------------------------------

	r := router.New(cfg, authHandler, jwtManager)

	// --------------------------------------------------
	// 5. HTTP server
	// --------------------------------------------------

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,

		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// --------------------------------------------------
	// 6. Start server
	// --------------------------------------------------

	go func() {
		log.Printf(
			"auth service running on port %s",
			cfg.Port,
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// --------------------------------------------------
	// 7. Graceful shutdown
	// --------------------------------------------------

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-stop

	log.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	log.Println("auth service stopped")
}
