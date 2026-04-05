package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newConfigDeleteCmd() *cobra.Command {
	return &cobra.Command{
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
}
