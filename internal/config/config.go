// File: internal/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ConfigFileName = "config.json"
	ConfigDirName  = "synkronus"
)

type GCPConfig struct {
	Project string `json:"project,omitempty"`
}

type AWSConfig struct {
	Region string `json:"region,omitempty"`
}

type Config struct {
	GCP *GCPConfig `json:"gcp,omitempty"`
	AWS *AWSConfig `json:"aws,omitempty"`
}

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

func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if len(data) == 0 {
		return &Config{}, nil
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
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

	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid config key format: %s. Use format like 'provider.key' (e.g., 'gcp.project')", key)
	}
	provider := parts[0]
	field := parts[1]

	switch provider {
	case "gcp":
		if config.GCP == nil {
			config.GCP = &GCPConfig{}
		}
		if field == "project" {
			config.GCP.Project = value
		} else {
			return fmt.Errorf("unknown config key for gcp: %s", field)
		}
	case "aws":
		if config.AWS == nil {
			config.AWS = &AWSConfig{}
		}
		if field == "region" {
			config.AWS.Region = value
		} else {
			return fmt.Errorf("unknown config key for aws: %s", field)
		}
	default:
		return fmt.Errorf("unknown provider in config key: %s", provider)
	}

	return SaveConfig(config)
}

func GetValue(key string) (string, bool, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", false, err
	}

	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return "", false, fmt.Errorf("invalid config key format: %s", key)
	}
	provider := parts[0]
	field := parts[1]

	switch provider {
	case "gcp":
		if config.GCP != nil && field == "project" {
			return config.GCP.Project, true, nil
		}
	case "aws":
		if config.AWS != nil && field == "region" {
			return config.AWS.Region, true, nil
		}
	}

	return "", false, nil
}

func DeleteValue(key string) (bool, error) {
	config, err := LoadConfig()
	if err != nil {
		return false, err
	}

	val, exists, err := GetValue(key)
	if err != nil {
		return false, err
	}
	if !exists || val == "" {
		return false, nil
	}

	parts := strings.SplitN(key, ".", 2)
	provider := parts[0]
	field := parts[1]

	switch provider {
	case "gcp":
		if config.GCP != nil && field == "project" {
			config.GCP.Project = ""
		}
	case "aws":
		if config.AWS != nil && field == "region" {
			config.AWS.Region = ""
		}
	}

	if err := SaveConfig(config); err != nil {
		return false, err
	}

	return true, nil
}
