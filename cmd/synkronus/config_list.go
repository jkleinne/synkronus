package main

import (
	"fmt"
	"maps"
	"slices"
	synkconfig "synkronus/internal/config"

	"github.com/spf13/cobra"
)

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
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
}
