// File: cmd/synkronus/config_cmd.go
package main

import (
	"fmt"
	"sort"
	"strings"

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
				return fmt.Errorf("error setting configuration: %v", err)
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
				return fmt.Errorf("error deleting configuration: %v", err)
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
			flattenedSettings := flattenConfigMap(settings)

			var displaySettings = make(map[string]interface{})
			for k, v := range flattenedSettings {
				if s, ok := v.(string); ok {
					if s != "" {
						displaySettings[k] = v
					}
				} else if v != nil {
					displaySettings[k] = v
				}
			}

			if len(displaySettings) == 0 {
				fmt.Println("No configuration values set. Use 'synkronus config set <key> <value>'.")
				return nil
			}

			keys := make([]string, 0, len(displaySettings))
			for k := range displaySettings {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			fmt.Println("Current configuration:")
			for _, k := range keys {
				fmt.Printf("  %s = %v\n", k, displaySettings[k])
			}

			return nil
		},
	}

	configCmd.AddCommand(configSetCmd, configGetCmd, configDeleteCmd, configListCmd)
	return configCmd
}

// Recursively flattens a nested map (like Viper's config) into a flat map with dot notation keys
func flattenConfigMap(nestedMap map[string]interface{}) map[string]interface{} {
	flattenedMap := make(map[string]interface{})

	var flatten func(string, interface{})
	flatten = func(prefix string, value interface{}) {
		switch v := value.(type) {
		case map[string]interface{}:
			for k, val := range v {
				newPrefix := k
				if prefix != "" {
					newPrefix = prefix + "." + k
				}
				flatten(newPrefix, val)
			}
		default:
			if prefix != "" {
				flattenedMap[prefix] = value
			}
		}
	}

	flatten("", nestedMap)
	return flattenedMap
}
