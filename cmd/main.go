package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"net/http"

	"github.com/Nzyazin/itk/internal/core/logger"
	"github.com/Nzyazin/itk/internal/server"
)

func main() {
	log, cleanup := logger.NewLogger()
	defer cleanup()

	srv, err := server.NewServer(log)
	if err != nil {
		log.Error("Failed to create server", logger.ErrorField("error", err))
		return
	}

	go func() {
		log.Info("Starting server", logger.StringField("port", ":8080"))
		if err := srv.Run(":8080"); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", logger.ErrorField("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server shutdown failed", logger.ErrorField("error", err))
	}

	log.Info("Server exited properly")
}