package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"synkronus/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  `Manage configuration settings for providers like GCP and AWS. You can set, get, list, and delete configuration values.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration key-value pair",
	Long:  `Sets a configuration value. For example: 'synkronus config set gcp_project my-gcp-123'`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		if err := config.SetValue(key, value); err != nil {
			fmt.Printf("Error setting configuration: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Configuration set: %s = %s\n", key, value)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value by key",
	Long:  `Retrieves a configuration value for a given key.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a configuration value by key",
	Long:  `Deletes a configuration value for a given key.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all current configuration values",
	Long:  `Displays all the key-value pairs currently stored in the configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configListCmd)
}
