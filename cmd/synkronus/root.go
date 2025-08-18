// File: cmd/synkronus/root.go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Creates the root command and wires up its subcommands using
// the application container for dependency injection
func newRootCmd(app *appContainer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "synkronus",
		Short: "Synkronus is a command-line tool for managing cloud resources.",
		Long: `A unified CLI to interact with various cloud services for resources
like storage, SQL databases, and more. Configure your providers and
manage your infrastructure from one place.`,
	}

	cmd.AddCommand(newStorageCmd(app))
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newSqlCmd())

	return cmd
}

func Execute(app *appContainer) {
	rootCmd := newRootCmd(app)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
