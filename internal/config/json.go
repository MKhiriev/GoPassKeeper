package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func parseJSON(jsonFilePath string) (*StructuredConfig, error) {
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading a json file: %w", err)
	}

	var jsonCfg StructuredConfig
	if err := json.NewDecoder(jsonFile).Decode(&jsonCfg); err != nil {
		return nil, fmt.Errorf("error decoding json configs: %w", err)
	}

	return &jsonCfg, nil
}
