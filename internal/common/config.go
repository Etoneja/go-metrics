package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/etoneja/go-metrics/internal/logger"
)

func LoadJSONConfig(cfg any, filePath string) error {
	if filePath == "" {
		logger.Get().Info("Config file path is empty, skipping JSON config load")
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
