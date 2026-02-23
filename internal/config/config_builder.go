// Package config provides application configuration loading and merging
// utilities for the go-pass-keeper application.
//
// Configuration is assembled from multiple sources in the following priority
// order (last source wins for non-zero fields):
//  1. Environment variables  — loaded via [withEnv]
//  2. Command-line flags     — loaded via [withFlags]
//  3. JSON file              — loaded via [withJSON], path resolved from the
//     sources above
//
// The entry point for production use is [GetStructuredConfig], which chains
// all three sources and validates the result.
package config

import (
	"errors"
	"fmt"

	"dario.cat/mergo"
)

// configBuilder accumulates partial [StructuredConfig] values from different
// sources and merges them into a single configuration on [build].
//
// The builder follows the fluent-interface pattern: each with* method appends
// a config source and returns the same *configBuilder so calls can be chained.
// Any error encountered during a with* step is stored in err and causes
// [build] to fail-fast without attempting to merge.
type configBuilder struct {
	// configs holds the ordered list of partial configurations to be merged.
	// Sources appended later take precedence over earlier ones for non-zero
	// fields (mergo.Merge semantics).
	configs []*StructuredConfig

	// err accumulates errors from individual source-loading steps.
	// Multiple errors are joined via errors.Join so all failures are visible
	// at once when build() is called.
	err error
}

// newConfigBuilder creates and returns an empty *configBuilder ready for use.
// The internal slice is pre-allocated for four sources (env, flags, JSON, and
// one spare) to avoid reallocations in the common case.
func newConfigBuilder() *configBuilder {
	return &configBuilder{
		configs: make([]*StructuredConfig, 0, 4),
	}
}

// build merges all accumulated partial configurations into a single
// [StructuredConfig] and validates the result.
//
// Merge order follows the order in which sources were appended: the first
// source provides the base, and each subsequent source fills in only the
// zero-value fields of the accumulator (mergo.Merge default behaviour).
//
// Returns an error if:
//   - any with* step previously recorded an error (b.err != nil);
//   - mergo.Merge fails for any source;
//   - the final config fails [StructuredConfig.validate].
func (b *configBuilder) build() (*StructuredConfig, error) {
	if b.err != nil {
		return nil, fmt.Errorf("error occured during building config: %w", b.err)
	}

	config := new(StructuredConfig)
	for _, cfg := range b.configs {
		if err := mergo.Merge(config, cfg); err != nil {
			return nil, fmt.Errorf("error merging configs: %w", err)
		}
	}

	return config, config.validate()
}

// withEnv parses environment variables into a [StructuredConfig] via
// [parseEnv] and appends the result to the builder.
//
// If parsing fails, the error is joined into b.err and the builder is
// returned unchanged so that subsequent steps are skipped gracefully.
//
// Returns the same *configBuilder to support method chaining.
func (b *configBuilder) withEnv() *configBuilder {
	envCfg := &StructuredConfig{}
	if err := parseEnv(envCfg); err != nil {
		b.err = errors.Join(b.err, err)
		return b
	}

	b.configs = append(b.configs, envCfg)
	return b
}

// withFlags parses command-line flags via [ParseFlags] and appends the
// resulting [StructuredConfig] to the builder.
//
// Flag parsing never returns an error directly; unknown flags are silently
// ignored. Returns the same *configBuilder to support method chaining.
func (b *configBuilder) withFlags() *configBuilder {
	flags := ParseFlags()

	b.configs = append(b.configs, flags)
	return b
}

// withJSON looks for a non-empty JSONFilePath field across all configs
// accumulated so far, and if found, parses that JSON file via [parseJSON],
// appending the result to the builder.
//
// When multiple sources specify a JSONFilePath, the last non-empty value
// wins. If no path is found, withJSON is a no-op.
//
// If parsing fails, the error is joined into b.err and the builder is
// returned unchanged. Returns the same *configBuilder to support method
// chaining.
func (b *configBuilder) withJSON() *configBuilder {
	var jsonPath string
	isJSONSpecified := false

	for _, cfg := range b.configs {
		if cfg.JSONFilePath != "" {
			isJSONSpecified = true
			jsonPath = cfg.JSONFilePath
		}
	}

	if isJSONSpecified {
		jsonCfg, err := parseJSON(jsonPath)
		if err != nil {
			b.err = errors.Join(b.err, err)
			return b
		}
		b.configs = append(b.configs, jsonCfg)
	}

	return b
}
