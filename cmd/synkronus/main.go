package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"synkronus/internal/config"
	"synkronus/pkg/formatter"
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
	var providerFlag string
	var remainingArgs []string

	if len(args) > 1 {
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--gcp":
				providerFlag = "gcp"
			case "--aws":
				providerFlag = "aws"
			default:
				remainingArgs = append(remainingArgs, args[i])
			}
		}
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	configMap := map[string]string{}
	for key, value := range cfg {
		if strValue, ok := value.(string); ok {
			configMap[key] = strValue
		}
	}

	ctx := context.Background()

	switch subcommand {
	case "list":
		handleStorageList(ctx, configMap, providerFlag)
	case "describe":
		if len(remainingArgs) < 1 {
			fmt.Println("Usage: synkronus storage describe <bucket-name> [--gcp|--aws]")
			fmt.Println("Note: Provider flag is required")
			os.Exit(1)
		}
		if providerFlag == "" {
			fmt.Println("Error: Provider flag (--gcp or --aws) is required for the describe subcommand")
			os.Exit(1)
		}
		handleStorageDescribe(ctx, configMap, providerFlag, remainingArgs[0])
	default:
		fmt.Printf("Unknown storage subcommand: %s\n", subcommand)
		fmt.Println("Available subcommands: list, describe")
		os.Exit(1)
	}
}

func initializeProvider(ctx context.Context, providerFlag string, configMap map[string]string) (storage.Storage, error) {
	switch providerFlag {
	case "gcp":
		gcpProject, hasProject := configMap["gcp_project"]
		if !hasProject || gcpProject == "" {
			return nil, fmt.Errorf("GCP project not configured. Use 'synkronus config set gcp_project <project-id>'")
		}
		return gcp.NewGCPStorage(ctx, gcpProject)
	case "aws":
		awsRegion, hasRegion := configMap["aws_region"]
		if !hasRegion || awsRegion == "" {
			return nil, fmt.Errorf("AWS region not configured. Use 'synkronus config set aws_region <region>'")
		}
		return aws.NewAWSStorage(awsRegion), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerFlag)
	}
}

func handleStorageList(ctx context.Context, configMap map[string]string, providerFlag string) {
	storageFormatter := formatter.NewStorageFormatter()
	var providersToQuery []string

	if providerFlag != "" {
		providersToQuery = append(providersToQuery, providerFlag)
	} else {
		if val, ok := configMap["gcp_project"]; ok && val != "" {
			providersToQuery = append(providersToQuery, "gcp")
		}
		if val, ok := configMap["aws_region"]; ok && val != "" {
			providersToQuery = append(providersToQuery, "aws")
		}
	}

	if len(providersToQuery) == 0 {
		fmt.Println("No providers configured or specified. Configure GCP/AWS using 'synkronus config set'.")
		return
	}

	var allBuckets []storage.Bucket
	var wg sync.WaitGroup

	type fetchResult struct {
		providerName string
		buckets      []storage.Bucket
		err          error
	}
	resultsChan := make(chan fetchResult, len(providersToQuery))

	for _, pName := range providersToQuery {
		wg.Add(1)
		go func(pName string) {
			defer wg.Done()

			client, err := initializeProvider(ctx, pName, configMap)
			if err != nil {
				resultsChan <- fetchResult{pName, nil, fmt.Errorf("initializing client: %w", err)}
				return
			}
			defer client.Close()

			buckets, err := client.ListBuckets(ctx)
			if err != nil {
				err = fmt.Errorf("listing buckets: %w", err)
			}
			resultsChan <- fetchResult{pName, buckets, err}
		}(pName)
	}

	wg.Wait()
	close(resultsChan)

	hasError := false
	for result := range resultsChan {
		if result.err != nil {
			fmt.Printf("Error fetching data from %s: %v\n", result.providerName, result.err)
			hasError = true
		} else {
			allBuckets = append(allBuckets, result.buckets...)
		}
	}

	if len(allBuckets) > 0 {
		fmt.Println(storageFormatter.FormatBucketList(allBuckets))
	} else if !hasError {
		fmt.Println("No buckets found.")
	}
}

func handleStorageDescribe(ctx context.Context, configMap map[string]string, providerFlag, bucketName string) {
	storageFormatter := formatter.NewStorageFormatter()

	client, err := initializeProvider(ctx, providerFlag, configMap)
	if err != nil {
		fmt.Printf("Error initializing provider: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	bucketDetails, err := client.DescribeBucket(ctx, bucketName)
	if err != nil {
		fmt.Printf("Error describing bucket '%s' on %s: %v\n", bucketName, providerFlag, err)
		os.Exit(1)
	}

	fmt.Println(storageFormatter.FormatBucketDetails(bucketDetails))
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
