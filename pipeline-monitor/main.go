package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pipeline-monitor/internal/app"
	"pipeline-monitor/internal/config"
)

func main() {
	cfg := config.Load()

	application := app.New(cfg)

	server := &http.Server{
		Addr:    cfg.Port,
		Handler: application.Router(),
	}

	gracefulShutdown(server, application)
}

func gracefulShutdown(server *http.Server, app *app.Application) {
	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the application (stop monitors, close DB connections)
	app.Shutdown(ctx)

	// Shutdown the HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
