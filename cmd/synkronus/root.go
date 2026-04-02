// File: cmd/synkronus/root.go
package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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

			// Redirect stderr away from the terminal — slog writes from
			// the service layer would corrupt Bubble Tea's alt-screen.
			// In debug mode, redirect to a log file; otherwise, discard.
			origStderr := os.Stderr
			var logWriter io.Writer = io.Discard
			if debugMode {
				logPath := filepath.Join(os.Getenv("HOME"), ".config", "synkronus", "debug.log")
				if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); err == nil {
					defer f.Close()
					logWriter = f
				}
			}
			os.Stderr = os.NewFile(0, os.DevNull)
			// Set a TUI-safe logger and override the default
			tuiLogger := slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{Level: slog.LevelDebug}))
			slog.SetDefault(tuiLogger)

			tuiErr := tui.Run(tui.Deps{
				StorageService: app.StorageService,
				SqlService:     app.SqlService,
				ConfigManager:  app.ConfigManager,
				Config:         app.Config,
				Factory:        app.ProviderFactory,
				Logger:         tuiLogger,
			})

			// Restore stderr after TUI exits
			os.Stderr = origStderr
			return tuiErr
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
