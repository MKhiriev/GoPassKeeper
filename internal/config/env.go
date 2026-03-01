// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// parseEnv populates cfg from environment variables using the caarlos0/env
// library. Struct fields are mapped via their `env` and `envPrefix` tags
// defined on [StructuredConfig] and its nested types.
//
// Returns a wrapped error if env.Parse fails (e.g. a required variable is
// missing or a value cannot be converted to the target type).
func parseEnv(cfg any) error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("error getting env configs: %w", err)
	}

	return nil
}
