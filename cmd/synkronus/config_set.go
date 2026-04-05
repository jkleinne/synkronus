package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
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
}
