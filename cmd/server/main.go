// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package main

import (
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/handler"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/server"
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

	log.Info().Msg("starting a server")
	log.Debug().Any("config", cfg).Msg("received configs")

	storages, err := store.NewStorages(cfg.Storage, log)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating storages")
	}

	services, err := service.NewServices(storages, cfg.App, log)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating services")
	}

	handlers, err := handler.NewHandlers(services, cfg.Server, log)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating handlers")
	}

	servers, err := server.NewServer(handlers, cfg.Server, log)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating server(s)")
	}

	servers.RunServer()
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
