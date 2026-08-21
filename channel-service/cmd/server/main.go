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

	"channel-service/internal/category"
	"channel-service/internal/channel"
	"channel-service/internal/config"
	"channel-service/internal/database"
	"channel-service/internal/follow"
	"channel-service/internal/router"
	"channel-service/internal/topic"
)

func main() {
	// --------------------------------------------------
	// 1. Load configuration
	// --------------------------------------------------

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf(
			"failed to load config: %v",
			err,
		)
	}

	// --------------------------------------------------
	// 2. Connect to PostgreSQL
	// --------------------------------------------------

	db, err := database.NewPostgres(
		cfg.DatabaseURL,
	)

	if err != nil {
		log.Fatalf(
			"failed to connect to database: %v",
			err,
		)
	}

	defer db.Close()

	log.Println("connected to PostgreSQL")

	// --------------------------------------------------
	// 3. Run migrations
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

	channelRepository := channel.NewRepository(db)

	categoryRepository := category.NewRepository(db)

	topicRepository := topic.NewRepository(db)

	followRepository := follow.NewRepository(db)

	// --------------------------------------------------
	// 5. Initialize services
	// --------------------------------------------------

	channelService := channel.NewService(
		channelRepository,
	)

	categoryService := category.NewService(
		categoryRepository,
	)

	topicService := topic.NewService(
		topicRepository,
	)

	followService := follow.NewService(
		followRepository,
	)

	// --------------------------------------------------
	// 6. Initialize handlers
	// --------------------------------------------------

	channelHandler := channel.NewHandler(
		channelService,
	)

	categoryHandler := category.NewHandler(
		categoryService,
	)

	topicHandler := topic.NewHandler(
		topicService,
	)

	followHandler := follow.NewHandler(
		followService,
	)

	// --------------------------------------------------
	// 7. Initialize router
	// --------------------------------------------------

	r := router.New(
		cfg,

		channelHandler,
		categoryHandler,
		topicHandler,
		followHandler,
	)

	// --------------------------------------------------
	// 8. HTTP Server
	// --------------------------------------------------

	server := &http.Server{
		Addr: ":" + cfg.Port,

		Handler: r,

		ReadTimeout: 15 * time.Second,

		WriteTimeout: 15 * time.Second,

		IdleTimeout: 60 * time.Second,
	}

	// --------------------------------------------------
	// 9. Start server
	// --------------------------------------------------

	go func() {
		log.Printf(
			"channel service running on port %s",
			cfg.Port,
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {

			log.Fatalf(
				"channel service failed: %v",
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
		"channel service stopped",
	)
}
