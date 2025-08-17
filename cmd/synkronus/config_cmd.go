package main

import (
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if err := config.SetValue(key, value); err != nil {
			return fmt.Errorf("error setting configuration: %v", err)
		}
		fmt.Printf("Configuration set: %s = %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value by key",
	Long:  `Retrieves a configuration value for a given key.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value, exists, err := config.GetValue(key)

		if err != nil {
			return fmt.Errorf("error getting configuration: %v", err)
		}

		if !exists {
			return fmt.Errorf("configuration key '%s' not found", key)
		}
		fmt.Printf("%s = %v\n", key, value)
		return nil
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a configuration value by key",
	Long:  `Deletes a configuration value for a given key.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		deleted, err := config.DeleteValue(key)

		if err != nil {
			return fmt.Errorf("error deleting configuration: %v", err)
		}

		if !deleted {
			return fmt.Errorf("configuration key '%s' not found", key)
		}
		fmt.Printf("Configuration key '%s' deleted\n", key)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all current configuration values",
	Long:  `Displays all the key-value pairs currently stored in the configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configValues, err := config.ListValues()
		if err != nil {
			return fmt.Errorf("error listing configuration: %v", err)
		}

		if len(configValues) == 0 {
			fmt.Println("No configuration values set")
			return nil
		}

		fmt.Println("Current configuration:")
		for key, value := range configValues {
			fmt.Printf("  %s = %v\n", key, value)
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configListCmd)
}
