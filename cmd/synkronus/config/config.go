package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const ConfigFileName = "synkronus.config.json"

type Config map[string]interface{}

func LoadConfig() (Config, error) {
	configPath := ConfigFileName

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}

func SaveConfig(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding config: %w", err)
	}

	if err := os.WriteFile(ConfigFileName, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

func SetValue(key, value string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config[key] = value

	return SaveConfig(config)
}

func ListValues() (Config, error) {
	return LoadConfig()
}
