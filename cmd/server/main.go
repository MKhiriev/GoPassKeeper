package main

import (
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	log := logger.NewLogger("go-pass-server")
	cfg, err := config.GetStructuredConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("error getting configs")
	}

	log.Debug().Any("config", cfg).Msg("received configs")

	storages, err := store.NewStorages(cfg.Storage)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating storages")
	}

	services, err := service.NewServices(cfg.Storage)
}

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}

	if buildDate == "" {
		buildDate = "N/A"
	}

	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
