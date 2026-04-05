// File: cmd/synkronus/config_cmd.go
package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	synkconfig "synkronus/internal/config"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
		Long:  `Manage configuration settings for providers. You can set, get, list, and delete configuration values.`,
	}

	configSetCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration key-value pair",
		Long:  `Sets a configuration value. For example: 'synkronus config set gcp.project my-gcp-123'`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			key := strings.ToLower(args[0])
			value := args[1]

			if err := app.ConfigManager.SetValue(key, value); err != nil {
				return fmt.Errorf("setting configuration %q: %w", key, err)
			}
			fmt.Printf("Configuration set: %s = %s\n", key, value)
			return nil
		},
	}

	configGetCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value by key",
		Long:  `Retrieves a configuration value for a given key. For example: 'synkronus config get gcp.project'`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			key := strings.ToLower(args[0])
			value, exists := app.ConfigManager.GetValue(key)

			if !exists || value == "" {
				return fmt.Errorf("configuration key '%s' not found or not set", key)
			}
			fmt.Printf("%s = %v\n", key, value)
			return nil
		},
	}

	configDeleteCmd := &cobra.Command{
		Use:   "delete [key]",
		Short: "Delete a configuration value by key",
		Long:  `Deletes a configuration value for a given key. For example: 'synkronus config delete gcp.project'`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			key := strings.ToLower(args[0])
			deleted, err := app.ConfigManager.DeleteValue(key)

			if err != nil {
				return fmt.Errorf("deleting configuration %q: %w", key, err)
			}

			if !deleted {
				return fmt.Errorf("configuration key '%s' not found", key)
			}
			fmt.Printf("Configuration key '%s' deleted\n", key)
			return nil
		},
	}

	configListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all current configuration values",
		Long:  `Displays all the key-value pairs currently stored in the configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			settings := app.ConfigManager.GetAllSettings()
			displaySettings := synkconfig.FlattenSettings(settings)
			for k, v := range displaySettings {
				if v == "" {
					delete(displaySettings, k)
				}
			}

			if len(displaySettings) == 0 {
				fmt.Println("No configuration values set. Use 'synkronus config set <key> <value>'.")
				return nil
			}

			keys := slices.Sorted(maps.Keys(displaySettings))

			fmt.Println("Current configuration:")
			for _, k := range keys {
				fmt.Printf("  %s = %s\n", k, displaySettings[k])
			}

			return nil
		},
	}

	configCmd.AddCommand(configSetCmd, configGetCmd, configDeleteCmd, configListCmd)
	return configCmd
}

