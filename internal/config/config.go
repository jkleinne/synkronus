// File: internal/config/config.go
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

const (
	ConfigFileName = "config.json"
	ConfigDirName  = "synkronus"
	// Ensures the directory is only accessible by the owner (rwx------)
	ConfigDirPermissions os.FileMode = 0700
	// Ensures the file is only accessible by the owner (rw-------)
	ConfigFilePermissions os.FileMode = 0600
)

type GCPConfig struct {
	Project string `json:"project,omitempty" validate:"required"`
}

type AWSConfig struct {
	Region   string `json:"region,omitempty" validate:"required"`
	Endpoint string `json:"endpoint,omitempty" validate:"omitempty,uri"`
}

type Config struct {
	GCP *GCPConfig `json:"gcp,omitempty" validate:"omitempty"`
	AWS *AWSConfig `json:"aws,omitempty" validate:"omitempty"`
}

// IsGCPConfigured returns true if the GCP configuration block is present
// and the project ID is set. Used by GCP provider registration callbacks.
func IsGCPConfigured(cfg *Config) bool {
	return cfg.GCP != nil && cfg.GCP.Project != ""
}

type ConfigManager struct {
	v         *viper.Viper
	validator *validator.Validate
}

func NewConfigManager() (*ConfigManager, error) {
	v := viper.New()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error determining user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", ConfigDirName)

	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath(configDir)
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist (first run), but other errors (e.g., parsing, permissions) must be reported
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	return &ConfigManager{
		v:         v,
		validator: validator.New(),
	}, nil
}

func (cm *ConfigManager) LoadConfig() (*Config, error) {
	var config Config
	if err := cm.unmarshalStrict(&config); err != nil {
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
	// Ensure the directory exists with secure permissions (0700)
	if err := os.MkdirAll(configDir, ConfigDirPermissions); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Write the configuration file
	if err := cm.v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	// Explicitly set secure permissions (0600) on the file itself (Defense-in-Depth)
	if err := os.Chmod(configPath, ConfigFilePermissions); err != nil {
		return fmt.Errorf("error setting secure permissions on config file: %w", err)
	}

	return nil
}

func (cm *ConfigManager) SetValue(key, value string) error {
	cm.v.Set(key, value)

	var config Config
	if err := cm.unmarshalStrict(&config); err != nil {
		if revertErr := cm.v.ReadInConfig(); revertErr != nil {
			return fmt.Errorf("%w (additionally, failed to revert config: %v)", err, revertErr)
		}
		return err
	}

	if err := cm.validateConfig(&config); err != nil {
		if revertErr := cm.v.ReadInConfig(); revertErr != nil {
			return fmt.Errorf("%w (additionally, failed to revert config: %v)", err, revertErr)
		}
		return err
	}

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

	var config Config
	if err := cm.unmarshalStrict(&config); err != nil {
		if revertErr := cm.v.ReadInConfig(); revertErr != nil {
			return false, fmt.Errorf("%w (additionally, failed to revert config: %v)", err, revertErr)
		}
		return false, fmt.Errorf("error parsing config after deletion: %w", err)
	}

	if err := cm.validateConfig(&config); err != nil {
		if revertErr := cm.v.ReadInConfig(); revertErr != nil {
			return false, fmt.Errorf("%w (additionally, failed to revert config: %v)", err, revertErr)
		}
		return false, fmt.Errorf("cannot delete key '%s': %w", key, err)
	}

	if err := cm.SaveConfig(); err != nil {
		return false, err
	}

	return true, nil
}

func (cm *ConfigManager) GetAllSettings() map[string]any {
	return cm.v.AllSettings()
}

// RemoveProvider removes an entire provider block (e.g., "gcp", "aws") from the
// configuration. This bypasses per-field validation since the entire block is
// removed — there are no dangling required fields. Returns true if the provider
// was configured and removed, false if it wasn't present.
func (cm *ConfigManager) RemoveProvider(providerName string) (bool, error) {
	providerName = strings.ToLower(providerName)

	settings := cm.v.AllSettings()
	if _, exists := settings[providerName]; !exists {
		return false, nil
	}

	// Remove the provider key from settings and write a fresh config file
	delete(settings, providerName)

	// Write the cleaned settings directly to file, bypassing Viper's
	// in-memory state which doesn't support key deletion.
	configPath, err := cm.getPreferredConfigPath()
	if err != nil {
		return false, err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, ConfigDirPermissions); err != nil {
		return false, fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := marshalJSON(settings)
	if err != nil {
		return false, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, ConfigFilePermissions); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}

	// Viper doesn't support key deletion — ReadInConfig merges with
	// existing state. Create a fresh Viper instance pointing at the same file.
	freshViper := viper.New()
	freshViper.SetConfigFile(configPath)
	if err := freshViper.ReadInConfig(); err != nil {
		return false, fmt.Errorf("failed to reload config after provider removal: %w", err)
	}
	cm.v = freshViper

	return true, nil
}

// marshalJSON encodes settings as indented JSON.
func marshalJSON(settings map[string]any) ([]byte, error) {
	return json.MarshalIndent(settings, "", "  ")
}

func (cm *ConfigManager) getPreferredConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", ConfigDirName)
	return filepath.Join(configDir, ConfigFileName), nil
}

func (cm *ConfigManager) unmarshalStrict(target any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:      target,
		ErrorUnused: true,
	})
	if err != nil {
		return fmt.Errorf("internal error: failed to create config decoder: %w", err)
	}

	if err := decoder.Decode(cm.v.AllSettings()); err != nil {
		if strings.Contains(err.Error(), "invalid keys") || strings.Contains(err.Error(), "unused keys") {
			return fmt.Errorf("unrecognized configuration key provided. Please use a valid key (e.g., 'gcp.project')")
		}
		return err
	}
	return nil
}

func (cm *ConfigManager) validateConfig(config *Config) error {
	err := cm.validator.Struct(config)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		var errs []string
		for _, fe := range validationErrors {
			namespace := strings.ToLower(fe.Namespace())
			errs = append(errs, fmt.Sprintf("field '%s' is invalid (rule: %s)", namespace, fe.Tag()))
		}
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errs, "; "))
	}

	return fmt.Errorf("invalid configuration: %w", err)
}
