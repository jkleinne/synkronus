// File: cmd/synkronus/root.go
package main

import (
	"fmt"
	"os"
	"synkronus/internal/flags"
	"synkronus/internal/output"
	"synkronus/internal/tui"

	"github.com/spf13/cobra"
)

// Creates the root command, defines global flags, and sets up the initialization hook
func newRootCmd() *cobra.Command {
	var debugMode bool
	var outputFormatStr string

	cmd := &cobra.Command{
		Use:   "synkronus",
		Short: "Synkronus is a command-line tool for managing cloud resources.",
		Long: `A unified CLI to interact with various cloud services for resources
like storage, SQL databases, and more. Configure your providers and
manage your infrastructure from one place.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Parse and validate the output format flag
			outputFormat, err := output.ParseFormat(outputFormatStr)
			if err != nil {
				return err
			}

			// Initialize the application container
			app, err := newApp(debugMode, outputFormat)
			if err != nil {
				return fmt.Errorf("failed to initialize application: %w", err)
			}

			if debugMode {
				app.Logger.Debug("Debug logging enabled")
			}

			// Inject the initialized container into the command's context
			// so subcommands can access it
			ctx := app.ToContext(cmd.Context())
			cmd.SetContext(ctx)

			return nil
		},
		// Launch TUI when no subcommand is given
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}
			return tui.Run(tui.Deps{
				StorageService: app.StorageService,
				SqlService:     app.SqlService,
				ConfigManager:  app.ConfigManager,
				Config:         app.Config,
				Factory:        app.ProviderFactory,
				Logger:         app.Logger,
			})
		},
		// Silence usage on error, error reporting is explicitly handled in Execute()
		SilenceUsage: true,
		// Silence errors so we don't print the error twice (once by Cobra, once Execute())
		SilenceErrors: true,
	}

	// Define persistent flags (available to all subcommands)
	cmd.PersistentFlags().BoolVarP(&debugMode, flags.Debug, flags.DebugShort, false, "Enable verbose debug logging")
	cmd.PersistentFlags().StringVarP(&outputFormatStr, flags.Output, flags.OutputShort, "table", "Output format: table, json, yaml")

	// Add subcommands
	cmd.AddCommand(newStorageCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newSqlCmd())

	return cmd
}

// Starts the CLI execution
func Execute() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
