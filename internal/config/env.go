package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

func parseEnv(cfg any) error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("error getting env configs: %w", err)
	}

	// todo implement method

	return nil
}
