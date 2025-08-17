// File: cmd/synkronus/root.go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "synkronus",
	Short: "Synkronus is a command-line tool for managing cloud resources.",
	Long: `A unified CLI to interact with various cloud services for resources
like storage, SQL databases, and more. Configure your providers and
manage your infrastructure from one place.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(storageCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(sqlCmd)
}
