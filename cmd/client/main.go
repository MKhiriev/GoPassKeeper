package main

import (
	"fmt"
	"os"

	"github.com/MKhiriev/go-pass-keeper/internal/client"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	app, err := client.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "init client app error: %v\n", err)
		os.Exit(1)
	}

	if err = app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "client run error: %v\n", err)
		os.Exit(1)
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
