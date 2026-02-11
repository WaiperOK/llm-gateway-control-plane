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

	"github.com/WaiperOK/llm-gateway-control-plane/internal/app"
	"github.com/WaiperOK/llm-gateway-control-plane/internal/config"
	"github.com/WaiperOK/llm-gateway-control-plane/internal/transport/httpapi"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	metrics := app.NewMetrics(prometheus.DefaultRegisterer)

	svc := app.NewService(cfg, logger, metrics, app.SimulatedModelClient{})
	handler := httpapi.NewHandler(logger, svc)

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("gateway starting", "addr", cfg.ListenAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown failed", "err", err)
		os.Exit(1)
	}
	logger.Info("gateway stopped")
}
