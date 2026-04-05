// File: cmd/synkronus/config_cmd.go
package main

import "github.com/spf13/cobra"

// newConfigCmd returns the "config" parent command with set/get/delete/list subcommands.
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
		Long:  `Manage configuration settings for providers. You can set, get, list, and delete configuration values.`,
	}
	cmd.AddCommand(newConfigSetCmd(), newConfigGetCmd(), newConfigDeleteCmd(), newConfigListCmd())
	return cmd
}
