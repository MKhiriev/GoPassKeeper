package main

import (
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/client"
	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/internal/tui"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	log := logger.NewClientLogger("go-pass-client")
	cfg, err := config.GetClientConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("error getting configs")
	}

	serverAdapter, err := adapter.NewHTTPServerAdapter(cfg.Adapter, cfg.App, log)
	if err != nil {
		log.Fatal().Err(err).Msg("create local adapter")
	}

	localStorage, err := store.NewClientStorages(cfg.Storage, log)
	if err != nil {
		log.Fatal().Err(err).Msg("create local storage")
	}

	services, err := service.NewClientServices(localStorage, serverAdapter, log)
	if err != nil {
		log.Fatal().Err(err).Msg("create client services")
	}

	ui, err := tui.New(services, log)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating ui")
	}

	app, err := client.NewApp(services, ui, cfg.Workers, log)
	if err != nil {
		log.Fatal().Err(err).Msg("init client app error")
	}

	if err = app.Run(); err != nil {
		log.Fatal().Err(err).Msg("client run error")
	}
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
