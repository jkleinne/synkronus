package main

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

func handleConfigCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Config command requires a subcommand")
		fmt.Println("Available subcommands: set, list")
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "set":
		handleConfigSetCommand(args[1:])
	case "list":
		handleConfigListCommand()
	default:
		fmt.Printf("Unknown config subcommand: %s\n", subcommand)
		fmt.Println("Available subcommands: set, list")
		os.Exit(1)
	}
}

func handleConfigSetCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: synkronus config set <key> <value>")
		fmt.Println("Example: synkronus config set gcp_project my-gcp-123")
		os.Exit(1)
	}

	key := args[0]
	value := args[1]

	if err := SetValue(key, value); err != nil {
		fmt.Printf("Error setting configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Configuration set: %s = %s\n", key, value)
}

func handleConfigListCommand() {
	configValues, err := ListValues()
	if err != nil {
		fmt.Printf("Error listing configuration: %v\n", err)
		os.Exit(1)
	}

	if len(configValues) == 0 {
		fmt.Println("No configuration values set")
		return
	}

	fmt.Println("Current configuration:")
	for key, value := range configValues {
		fmt.Printf("  %s = %v\n", key, value)
	}
}
