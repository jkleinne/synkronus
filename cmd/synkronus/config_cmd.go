package main

import (
	"fmt"
	"os"

	"synkronus/internal/config"
)

func handleConfigCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Config command requires a subcommand")
		fmt.Println("Available subcommands: get, set, delete, list")
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "set":
		handleConfigSetCommand(args[1:])
	case "get":
		handleConfigGetCommand(args[1:])
	case "delete":
		handleConfigDeleteCommand(args[1:])
	case "list":
		handleConfigListCommand()
	default:
		fmt.Printf("Unknown config subcommand: %s\n", subcommand)
		fmt.Println("Available subcommands: get, set, delete, list")
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

	if err := config.SetValue(key, value); err != nil {
		fmt.Printf("Error setting configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Configuration set: %s = %s\n", key, value)
}

func handleConfigListCommand() {
	configValues, err := config.ListValues()
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

func handleConfigGetCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: synkronus config get <key>")
		fmt.Println("Example: synkronus config get gcp_project")
		os.Exit(1)
	}

	key := args[0]
	value, exists, err := config.GetValue(key)

	if err != nil {
		fmt.Printf("Error getting configuration: %v\n", err)
		os.Exit(1)
	}

	if !exists {
		fmt.Printf("Configuration key '%s' not found\n", key)
		os.Exit(1)
	}

	fmt.Printf("%s = %v\n", key, value)
}

func handleConfigDeleteCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: synkronus config delete <key>")
		fmt.Println("Example: synkronus config delete gcp_project")
		os.Exit(1)
	}

	key := args[0]
	deleted, err := config.DeleteValue(key)

	if err != nil {
		fmt.Printf("Error deleting configuration: %v\n", err)
		os.Exit(1)
	}

	if !deleted {
		fmt.Printf("Configuration key '%s' not found\n", key)
		os.Exit(1)
	}

	fmt.Printf("Configuration key '%s' deleted\n", key)
}
