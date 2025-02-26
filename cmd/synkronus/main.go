package main

import (
	"fmt"
	"os"

	"synkronus/pkg/storage"
	"synkronus/pkg/storage/aws"
	"synkronus/pkg/storage/gcp"
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
		fmt.Println("Available subcommands: list, describe")
		os.Exit(1)
	}

	subcommand := args[0]
	var provider string
	var remainingArgs []string

	// Check for provider flags
	if len(args) > 1 {
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--gcp":
				provider = "gcp"
			case "--aws":
				provider = "aws"
			default:
				remainingArgs = append(remainingArgs, args[i])
			}
		}
	}

	// Load configuration for storage providers
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Convert Config to map[string]string for provider configuration
	configMap := map[string]string{}
	for key, value := range config {
		if strValue, ok := value.(string); ok {
			configMap[key] = strValue
		}
	}

	switch subcommand {
	case "list":
		handleStorageList(configMap, provider)
	case "describe":
		if len(remainingArgs) < 1 {
			fmt.Println("Usage: synkronus storage describe <bucket-name> [--gcp|--aws]")
			fmt.Println("Note: Provider flag is required")
			os.Exit(1)
		}
		if provider == "" {
			fmt.Println("Error: Provider flag (--gcp or --aws) is required for the describe subcommand")
			os.Exit(1)
		}
		handleStorageDescribe(configMap, provider, remainingArgs[0])
	default:
		fmt.Printf("Unknown storage subcommand: %s\n", subcommand)
		fmt.Println("Available subcommands: list, describe")
		os.Exit(1)
	}
}

func handleStorageList(configMap map[string]string, provider string) {
	// Use storage package to keep import
	_ = storage.Provider(provider)
	// If no provider specified, list all configured providers
	if provider == "" {
		fmt.Println("Listing storage buckets across all configured providers:")

		// Check if GCP is configured
		if gcpProject, hasProject := configMap["gcp_project"]; hasProject {
			fmt.Println("Provider: gcp")
			// Initialize GCP client without requiring a specific bucket
			gcpClient := gcp.NewGCPStorage(gcpProject, "")
			buckets, err := gcpClient.List()
			if err != nil {
				fmt.Printf("Error listing GCP buckets: %v\n", err)
			} else {
				for _, bucket := range buckets {
					fmt.Printf("  - %s\n", bucket)
				}
			}
		}

		// Check if AWS is configured
		if awsRegion, hasRegion := configMap["aws_region"]; hasRegion {
			fmt.Println("Provider: aws")
			// Initialize AWS client without requiring a specific bucket
			awsClient := aws.NewAWSStorage(awsRegion, "")
			buckets, err := awsClient.List()
			if err != nil {
				fmt.Printf("Error listing AWS buckets: %v\n", err)
			} else {
				for _, bucket := range buckets {
					fmt.Printf("  - %s\n", bucket)
				}
			}
		}
		return
	}

	// List buckets for specific provider
	fmt.Printf("Listing storage buckets for provider: %s\n", provider)

	switch provider {
	case "gcp":
		gcpProject, hasProject := configMap["gcp_project"]
		if !hasProject {
			fmt.Println("Error: GCP project not configured. Use 'synkronus config set gcp_project <project-id>'")
			return
		}
		gcpClient := gcp.NewGCPStorage(gcpProject, "")
		buckets, err := gcpClient.List()
		if err != nil {
			fmt.Printf("Error listing GCP buckets: %v\n", err)
			return
		}
		for _, bucket := range buckets {
			fmt.Printf("  - %s\n", bucket)
		}

	case "aws":
		awsRegion, hasRegion := configMap["aws_region"]
		if !hasRegion {
			fmt.Println("Error: AWS region not configured. Use 'synkronus config set aws_region <region>'")
			return
		}
		awsClient := aws.NewAWSStorage(awsRegion, "")
		buckets, err := awsClient.List()
		if err != nil {
			fmt.Printf("Error listing AWS buckets: %v\n", err)
			return
		}
		for _, bucket := range buckets {
			fmt.Printf("  - %s\n", bucket)
		}

	default:
		fmt.Printf("Unsupported provider: %s\n", provider)
	}
}

func handleStorageDescribe(configMap map[string]string, provider, bucketName string) {
	switch provider {
	case "gcp":
		gcpProject, hasProject := configMap["gcp_project"]
		if !hasProject {
			fmt.Println("Error: GCP project not configured. Use 'synkronus config set gcp_project <project-id>'")
			return
		}
		gcpClient := gcp.NewGCPStorage(gcpProject, bucketName)
		details, err := gcpClient.DescribeBucket(bucketName)
		if err != nil {
			fmt.Printf("Error describing GCP bucket: %v\n", err)
			return
		}
		fmt.Printf("Details for bucket '%s' on GCP:\n", bucketName)
		for key, value := range details {
			fmt.Printf("  %s: %s\n", key, value)
		}

	case "aws":
		awsRegion, hasRegion := configMap["aws_region"]
		if !hasRegion {
			fmt.Println("Error: AWS region not configured. Use 'synkronus config set aws_region <region>'")
			return
		}
		awsClient := aws.NewAWSStorage(awsRegion, bucketName)
		details, err := awsClient.DescribeBucket(bucketName)
		if err != nil {
			fmt.Printf("Error describing AWS bucket: %v\n", err)
			return
		}
		fmt.Printf("Details for bucket '%s' on AWS:\n", bucketName)
		for key, value := range details {
			fmt.Printf("  %s: %s\n", key, value)
		}

	default:
		fmt.Printf("Unsupported provider: %s\n", provider)
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
	fmt.Println("\nStorage Subcommands:")
	fmt.Println("  list       List storage buckets (all providers or specific with --gcp/--aws flag)")
	fmt.Println("  describe   Get details of a specific bucket (requires bucket name and provider flag)")
	fmt.Println("\nExamples:")
	fmt.Println("  synkronus storage list")
	fmt.Println("  synkronus storage list --gcp")
	fmt.Println("  synkronus storage describe exampleBucket --gcp")
	fmt.Println("  synkronus config get gcp_project")
	fmt.Println("  synkronus config set gcp_project my-gcp-123")
}
