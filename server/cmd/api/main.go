package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"poker_server/internal/ws"
)

type Config struct {
	Port           string
	AllowedOrigins []string
	Environment    string /* "development" or "production" */
}

func loadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}

	return Config {
		Port:           port,
		Environment:    env,
		AllowedOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
	}
}

func main() {
	cfg := loadConfig()

	var logger *slog.Logger
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	hub := ws.NewHub()
	go hub.Run(ctx)

	r := gin.New()
	r.Use(gin.Recovery()) // Prevents full crash and makes the server return 500 Internal Server Error

	wsHandler := ws.NewHandler(hub, cfg.AllowedOrigins, logger)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/ws", wsHandler.Serve)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Port, "env", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced server shutdown", "error", err)
	}

	logger.Info("server exited cleanly")
}