// File: internal/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
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

// ConfigManager encapsulates all configuration logic, managing its own viper instance
type ConfigManager struct {
	v *viper.Viper
}

// Creates and initializes a new ConfigManager
// It sets up a new viper instance, defines config paths, and reads the configuration from the file
func NewConfigManager() (*ConfigManager, error) {
	v := viper.New()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// This is unlikely to fail, but if it does, Viper will only look in the current directory
	}

	configDir := filepath.Join(homeDir, ".config", ConfigDirName)

	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath(configDir)
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// A real error occurred (e.g., malformed JSON)
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	return &ConfigManager{v: v}, nil
}

func (cm *ConfigManager) LoadConfig() (*Config, error) {
	var config Config
	if err := cm.v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}
	return &config, nil
}

func (cm *ConfigManager) SaveConfig() error {
	configPath, err := cm.getPreferredConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Use Viper's atomic WriteConfigAs to prevent corruption
	if err := cm.v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

func (cm *ConfigManager) SetValue(key, value string) error {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid config key format: %s. Use format like 'provider.key' (e.g., 'gcp.project')", key)
	}

	provider := parts[0]
	field := parts[1]

	// Validate the key to prevent arbitrary keys from being set
	switch provider {
	case "gcp":
		if field != "project" {
			return fmt.Errorf("unknown config key for gcp: %s", field)
		}
	case "aws":
		if field != "region" {
			return fmt.Errorf("unknown config key for aws: %s", field)
		}
	default:
		return fmt.Errorf("unknown provider in config key: %s", provider)
	}

	cm.v.Set(key, value)
	return cm.SaveConfig()
}

func (cm *ConfigManager) GetValue(key string) (string, bool) {
	if !cm.v.IsSet(key) {
		return "", false
	}
	value := cm.v.GetString(key)
	return value, value != ""
}

func (cm *ConfigManager) DeleteValue(key string) (bool, error) {
	val, exists := cm.GetValue(key)
	if !exists || val == "" {
		return false, nil
	}

	cm.v.Set(key, "")
	if err := cm.SaveConfig(); err != nil {
		return false, err
	}
	return true, nil
}

func (cm *ConfigManager) GetAllSettings() map[string]interface{} {
	return cm.v.AllSettings()
}

func (cm *ConfigManager) getPreferredConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", ConfigDirName)
	return filepath.Join(configDir, ConfigFileName), nil
}
