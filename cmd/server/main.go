package main

import "fmt"

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()
	fmt.Println("Server started!")
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
