package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigFileName = "config.json"
	ConfigDirName  = "synkronus"
)

type Config map[string]interface{}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", ConfigDirName)
	configPath := filepath.Join(configDir, ConfigFileName)

	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	if _, err := os.Stat(ConfigFileName); err == nil {
		if err := migrateConfig(ConfigFileName, configPath); err == nil {
			return configPath, nil
		}
		return ConfigFileName, nil
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("error creating config directory: %w", err)
	}

	return configPath, nil
}

func migrateConfig(sourcePath, destPath string) error {
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("error reading source config file: %w", err)
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("error writing destination config file: %w", err)
	}

	return nil
}

func LoadConfig() (Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

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
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
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

func GetValue(key string) (interface{}, bool, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, false, err
	}

	value, exists := config[key]
	return value, exists, nil
}

func DeleteValue(key string) (bool, error) {
	config, err := LoadConfig()
	if err != nil {
		return false, err
	}

	if _, exists := config[key]; !exists {
		return false, nil
	}

	delete(config, key)

	if err := SaveConfig(config); err != nil {
		return false, err
	}

	return true, nil
}
