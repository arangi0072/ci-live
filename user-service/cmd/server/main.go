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

	"user-service/internal/config"
	"user-service/internal/database"
	"user-service/internal/device"
	"user-service/internal/router"
	"user-service/internal/user"
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
	// 3. Run database migrations
	// --------------------------------------------------

	migrationCtx, migrationCancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)

	defer migrationCancel()

	if err := database.RunMigrations(
		migrationCtx,
		db,
		cfg.MigrationsPath,
	); err != nil {
		log.Fatalf(
			"failed to run migrations: %v",
			err,
		)
	}

	log.Println("database migrations completed")

	// --------------------------------------------------
	// 4. Initialize repositories
	// --------------------------------------------------

	userRepository := user.NewRepository(db)

	deviceRepository := device.NewRepository(db)

	// --------------------------------------------------
	// 5. Initialize services
	// --------------------------------------------------

	userService := user.NewService(
		userRepository,
	)

	deviceService := device.NewService(
		deviceRepository,
	)

	// --------------------------------------------------
	// 6. Initialize handlers
	// --------------------------------------------------

	userHandler := user.NewHandler(
		userService,
	)

	deviceHandler := device.NewHandler(
		deviceService,
	)

	// --------------------------------------------------
	// 7. Initialize router
	// --------------------------------------------------

	r := router.New(
		cfg,
		userHandler,
		deviceHandler,
	)

	// --------------------------------------------------
	// 8. HTTP server
	// --------------------------------------------------

	server := &http.Server{
		Addr: ":" + cfg.Port,

		Handler: r,

		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// --------------------------------------------------
	// 9. Start HTTP server
	// --------------------------------------------------

	go func() {
		log.Printf(
			"user service running on port %s",
			cfg.Port,
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {

			log.Fatalf(
				"user service failed: %v",
				err,
			)
		}
	}()

	// --------------------------------------------------
	// 10. Wait for shutdown signal
	// --------------------------------------------------

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-stop

	log.Println(
		"shutdown signal received",
	)

	// --------------------------------------------------
	// 11. Graceful shutdown
	// --------------------------------------------------

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer shutdownCancel()

	if err := server.Shutdown(
		shutdownCtx,
	); err != nil {

		log.Printf(
			"graceful shutdown failed: %v",
			err,
		)
	}

	log.Println(
		"user service stopped",
	)
}
