package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "storage":
		handleStorageCommand(os.Args[2:])
	case "sql":
		handleSQLCommand(os.Args[2:])
	case "config":
		handleConfigCommand(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleStorageCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Storage command requires a subcommand")
		fmt.Println("Available subcommands: list")
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "list":
		fmt.Println("Listing storage...")
		// TODO: Implement buckets listing functionality (GCP for now)
	default:
		fmt.Printf("Unknown storage subcommand: %s\n", subcommand)
		fmt.Println("Available subcommands: list")
		os.Exit(1)
	}
}

func handleSQLCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("SQL command requires a subcommand")
		fmt.Println("Available subcommands: list")
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "list":
		fmt.Println("Listing SQL resources...")
		// TODO: Implement SQL listing functionality (GCP for now)
	default:
		fmt.Printf("Unknown SQL subcommand: %s\n", subcommand)
		fmt.Println("Available subcommands: list")
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Synkronus - Command Line Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  synkronus <command> <subcommand> [options]")
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  storage    Manage storage resources")
	fmt.Println("  sql        Manage SQL resources")
	fmt.Println("  config     Manage configuration settings")
	fmt.Println("  help       Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  synkronus storage list")
	fmt.Println("  synkronus sql list")
	fmt.Println("  synkronus config get gcp_project")
	fmt.Println("  synkronus config set gcp_project my-gcp-123")
	fmt.Println("  synkronus config delete gcp_project")
	fmt.Println("  synkronus config list")
}
