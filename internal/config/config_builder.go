package config

import (
	"errors"
	"fmt"

	"dario.cat/mergo"
)

type configBuilder struct {
	configs []*StructuredConfig
	err     error
}

func newConfigBuilder() *configBuilder {
	return &configBuilder{
		configs: make([]*StructuredConfig, 0, 4),
	}
}

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

func (b *configBuilder) withEnv() *configBuilder {
	envCfg := &StructuredConfig{}
	if err := parseEnv(envCfg); err != nil {
		b.err = errors.Join(b.err, err)
		return b
	}

	b.configs = append(b.configs, envCfg)
	return b
}

func (b *configBuilder) withFlags() *configBuilder {
	flags := ParseFlags()

	b.configs = append(b.configs, flags)
	return b
}

func (b *configBuilder) withJSON() *configBuilder {
	var jsonPath string
	isJSONSpecified := false

	for _, cfg := range b.configs {
		if cfg.jsonFilePath != "" {
			isJSONSpecified = true
			jsonPath = cfg.jsonFilePath
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
