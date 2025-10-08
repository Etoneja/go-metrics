package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func LoadJSONConfig(cfg any, filePath string) error {
	if filePath == "" {
		log.Printf("Config file path is empty, skipping JSON config load")
		return nil
	}
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read config file %s: %w", filePath, err)
	}

	if err := json.Unmarshal(fileData, cfg); err != nil {
		return fmt.Errorf("cannot parse config file %s: %w", filePath, err)
	}

	return nil
}
