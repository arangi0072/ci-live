package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"stream-service/internal/chat"
	"stream-service/internal/config"
	"stream-service/internal/database"
	"stream-service/internal/endpoint"
	"stream-service/internal/like"
	"stream-service/internal/recording"
	"stream-service/internal/router"
	"stream-service/internal/session"
	"stream-service/internal/stream"
)

func main() {
	//
	// Load configuration
	//

	cfg, err := config.Load()
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}

	//
	// Initialize logger
	//

	logger, err := initLogger(cfg.App.Environment)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	defer func() {
		_ = logger.Sync()
	}()

	logger.Info(
		"starting CI Live server",
		zap.String("environment", cfg.App.Environment),
		zap.String("version", cfg.App.Version),
	)

	//
	// Root context
	//

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//
	// Handle OS shutdown signals
	//

	signalCtx, stopSignal := signal.NotifyContext(
		ctx,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stopSignal()

	//
	// PostgreSQL
	//

	db, err := database.NewPostgresPool(
		signalCtx,
		cfg.Postgres,
	)
	if err != nil {
		logger.Fatal(
			"failed to connect to PostgreSQL",
			zap.Error(err),
		)
	}

	defer db.Close()

	logger.Info("PostgreSQL connection established")

	//
	// Database migrations
	//

	if cfg.App.Environment != "test" {
		if err := database.RunMigration(
			signalCtx,
			db,
			"migrations/001_init.sql",
		); err != nil {
			logger.Fatal(
				"failed to run database migrations",
				zap.Error(err),
			)
		}

		logger.Info("database migrations completed")
	}

	//
	// =========================
	// REPOSITORIES
	// =========================
	//

	streamRepository := stream.NewRepository(db)
	sessionRepository := session.NewRepository(db)
	endpointRepository := endpoint.NewRepository(db)
	recordingRepository := recording.NewRepository(db)
	chatRepository := chat.NewRepository(db)
	likeRepository := like.NewRepository(db)

	//
	// =========================
	// SERVICES
	// =========================
	//

	streamService := stream.NewService(
		streamRepository,
	)

	sessionService := session.NewService(
		sessionRepository,
	)

	endpointService := endpoint.NewService(
		endpointRepository,
	)

	recordingService := recording.NewService(
		recordingRepository,
	)

	chatService := chat.NewService(
		chatRepository,
	)

	likeService := like.NewService(
		likeRepository,
	)

	//
	// =========================
	// HANDLERS
	// =========================
	//

	streamHandler := stream.NewHandler(
		streamService,
	)

	sessionHandler := session.NewHandler(
		sessionService,
	)

	endpointHandler := endpoint.NewHandler(
		endpointService,
	)

	recordingHandler := recording.NewHandler(
		recordingService,
	)

	chatHandler := chat.NewHandler(
		chatService,
	)

	likeHandler := like.NewHandler(
		likeService,
	)

	//
	// =========================
	// ROUTER
	// =========================
	//

	engine := router.NewRouter(
		router.Dependencies{
			Logger:           logger,
			ChatHandler:      chatHandler,
			EndpointHandler:  endpointHandler,
			LikeHandler:      likeHandler,
			RecordingHandler: recordingHandler,
			SessionHandler:   sessionHandler,
			StreamHandler:    streamHandler,
		},
	)

	//
	// HTTP server
	//

	server := &http.Server{
		Addr:         cfg.ServerAddress(),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	//
	// Start HTTP server
	//

	serverErr := make(chan error, 1)

	go func() {
		logger.Info(
			"HTTP server listening",
			zap.String("address", cfg.ServerAddress()),
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	//
	// Wait for shutdown signal or server failure
	//

	select {
	case err := <-serverErr:
		logger.Error(
			"HTTP server stopped unexpectedly",
			zap.Error(err),
		)

	case <-signalCtx.Done():
		logger.Info("shutdown signal received")
	}

	//
	// Graceful shutdown
	//

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(
			"graceful HTTP shutdown failed",
			zap.Error(err),
		)

		_ = server.Close()
	}

	logger.Info("CI Live server stopped")
}

//
// Logger
//

func initLogger(environment string) (*zap.Logger, error) {
	if environment == "production" {
		return zap.NewProduction()
	}

	return zap.NewDevelopment()
}
