package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deanandreas/ecommerce-api/internal/server"
)

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	app, err := server.NewAPI()
	if err != nil {
		slog.Error("server error", "error", err.Error())
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server start...")
		if err := app.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to run server", "error", err.Error())
			return
		}
		stop()
	}()

	<-ctx.Done()

	slog.Info("server shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err.Error())
		return
	}

	slog.Info("server shutdown gracefully")
}
