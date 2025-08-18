// File: cmd/synkronus/config_cmd.go
package main

import (
	"fmt"
	"strings"

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
	Long:  `Sets a configuration value. For example: 'synkronus config set gcp.project my-gcp-123'`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])
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
	Long:  `Retrieves a configuration value for a given key. For example: 'synkronus config get gcp.project'`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])
		value, exists, err := config.GetValue(key)

		if err != nil {
			return fmt.Errorf("error getting configuration: %v", err)
		}

		if !exists || value == "" {
			return fmt.Errorf("configuration key '%s' not found or not set", key)
		}
		fmt.Printf("%s = %v\n", key, value)
		return nil
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a configuration value by key",
	Long:  `Deletes a configuration value for a given key. For example: 'synkronus config delete gcp.project'`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])
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
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("error listing configuration: %v", err)
		}

		hasValues := false
		var output strings.Builder

		if cfg.GCP != nil && cfg.GCP.Project != "" {
			output.WriteString(fmt.Sprintf("  gcp.project = %s\n", cfg.GCP.Project))
			hasValues = true
		}
		if cfg.AWS != nil && cfg.AWS.Region != "" {
			output.WriteString(fmt.Sprintf("  aws.region = %s\n", cfg.AWS.Region))
			hasValues = true
		}

		if !hasValues {
			fmt.Println("No configuration values set. Use 'synkronus config set <key> <value>'.")
			return nil
		}

		fmt.Println("Current configuration:")
		fmt.Print(output.String())
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configListCmd)
}
