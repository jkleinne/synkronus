package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
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
}
