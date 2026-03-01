// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package config

import "strings"

// validate checks that the final merged [StructuredConfig] satisfies all
// application invariants before it is used at startup.
//
// Currently a no-op placeholder; validation rules will be added as the
// application matures (e.g. requiring non-empty DSN, token sign key, etc.).
//
// Returns nil if the configuration is valid, or a descriptive error otherwise.
func (cfg *StructuredConfig) validate() error {
	return nil
}

func (cfg *ClientConfig) validate() error {
	if cfg.Storage.DB.DSN == "" || strings.Contains(cfg.Storage.DB.DSN, "memory") {
		return ErrInvalidStorageConfigs
	}

	if cfg.Adapter.HTTPAddress == "" || cfg.Adapter.RequestTimeout == 0 {
		return ErrInvalidAdapterConfigs
	}

	if cfg.Workers.SyncInterval == 0 {
		return ErrInvalidWorkerConfigs
	}

	if cfg.App.HashKey == "" {
		return ErrInvalidAppConfigs
	}

	return nil
}
