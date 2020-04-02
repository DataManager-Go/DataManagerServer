package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JojiiOfficial/DataManagerServer/services"
	"github.com/jinzhu/gorm"

	log "github.com/sirupsen/logrus"
)

// Services
var (
	apiService *services.APIService // Handle endpoints
)

func startAPI() {
	log.Info("Starting version " + version)

	// Create the APIService and start it
	apiService = services.NewAPIService(config, db)
	apiService.Start()

	// Startup done
	log.Info("Startup completed")

	// Start loop to tick the services
	go (func() {
		for {
			time.Sleep(time.Hour)

		}
	})()

	awaitExit(apiService, db)
}

// Shutdown server gracefully
func awaitExit(httpServer *services.APIService, db *gorm.DB) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	// await os signal
	<-signalChan

	// Create a deadline for the await
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	log.Info("Shutting down server")

	if httpServer.HTTPServer != nil {
		err := httpServer.HTTPServer.Shutdown(ctx)
		if err != nil {
			log.Warn(err)
		}
		log.Info("HTTP server shutdown complete")
	}

	if httpServer.HTTPTLSServer != nil {
		err := httpServer.HTTPTLSServer.Shutdown(ctx)
		if err != nil {
			log.Warn(err)
		}
		log.Info("HTTPs server shutdown complete")
	}

	// Close db connection
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Warn(err)
		}
		log.Info("Database shutdown complete")
	}

	log.Info("Shutting down complete")
	os.Exit(0)
}
