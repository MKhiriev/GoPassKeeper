// Package config provides configuration loading, merging, and validation
// facilities for the application.
//
// Configuration is assembled from multiple sources in the following priority
// order (later sources override earlier non-zero fields):
//  1. Environment variables
//  2. Command-line flags
//  3. JSON config file
//
// The main entry points are [GetStructuredConfig] for server/runtime
// configuration and [GetClientConfig] for client-specific configuration.
package config
