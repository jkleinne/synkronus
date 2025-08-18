// File: cmd/synkronus/storage_cmd.go
package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"synkronus/internal/config"
	"synkronus/pkg/formatter"
	"synkronus/pkg/storage"
	"synkronus/pkg/storage/aws"
	"synkronus/pkg/storage/gcp"
)

var (
	gcpProvider bool
	awsProvider bool
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage storage resources like buckets",
	Long:  `The storage command allows you to list and describe storage buckets from configured cloud providers like AWS and GCP.`,
}

var storageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storage buckets",
	Long:  `Lists all storage buckets from the configured cloud providers. Use flags to specify a provider.`,
	RunE:  runStorageList,
}

var storageDescribeCmd = &cobra.Command{
	Use:   "describe [bucket-name]",
	Short: "Describe a specific storage bucket",
	Long:  `Provides detailed information about a specific storage bucket. You must specify the bucket name and the provider flag (--gcp or --aws).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runStorageDescribe,
}

func init() {
	storageCmd.AddCommand(storageListCmd)
	storageCmd.AddCommand(storageDescribeCmd)

	storageCmd.PersistentFlags().BoolVar(&gcpProvider, "gcp", false, "Use GCP provider")
	storageCmd.PersistentFlags().BoolVar(&awsProvider, "aws", false, "Use AWS provider")
}

func initializeProvider(ctx context.Context, providerFlag string, cfg *config.Config) (storage.Storage, error) {
	switch providerFlag {
	case "gcp":
		if cfg.GCP == nil || cfg.GCP.Project == "" {
			return nil, fmt.Errorf("GCP project not configured. Use 'synkronus config set gcp.project <project-id>'")
		}
		return gcp.NewGCPStorage(ctx, cfg.GCP.Project)
	case "aws":
		if cfg.AWS == nil || cfg.AWS.Region == "" {
			return nil, fmt.Errorf("AWS region not configured. Use 'synkronus config set aws.region <region>'")
		}
		return aws.NewAWSStorage(cfg.AWS.Region), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerFlag)
	}
}

func runStorageList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	storageFormatter := formatter.NewStorageFormatter()
	var providersToQuery []string

	// Determine which providers to query based on flags or configuration
	onlyGCP := gcpProvider && !awsProvider
	onlyAWS := awsProvider && !gcpProvider
	noFlags := !gcpProvider && !awsProvider

	if onlyGCP {
		providersToQuery = append(providersToQuery, "gcp")
	} else if onlyAWS {
		providersToQuery = append(providersToQuery, "aws")
	} else { // both flags or no flags
		gcpConfigured := cfg.GCP != nil && cfg.GCP.Project != ""
		awsConfigured := cfg.AWS != nil && cfg.AWS.Region != ""

		if (gcpConfigured && noFlags) || gcpProvider {
			providersToQuery = append(providersToQuery, "gcp")
		}
		if (awsConfigured && noFlags) || awsProvider {
			providersToQuery = append(providersToQuery, "aws")
		}
	}

	if len(providersToQuery) == 0 {
		fmt.Println("No providers configured or specified. Configure providers using 'synkronus config set'.")
		return nil
	}

	var allBuckets []storage.Bucket
	var mu sync.Mutex
	g, ctx := errgroup.WithContext(cmd.Context())

	for _, pName := range providersToQuery {
		// Capture pName for the goroutine
		pName := pName
		g.Go(func() error {
			client, err := initializeProvider(ctx, pName, cfg)
			if err != nil {
				return fmt.Errorf("initializing client for %s: %w", pName, err)
			}
			defer client.Close()

			buckets, err := client.ListBuckets(ctx)
			if err != nil {
				return fmt.Errorf("listing buckets from %s: %w", pName, err)
			}

			mu.Lock()
			allBuckets = append(allBuckets, buckets...)
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if len(allBuckets) > 0 {
		fmt.Println(storageFormatter.FormatBucketList(allBuckets))
	} else {
		fmt.Println("No buckets found.")
	}

	return nil
}

func runStorageDescribe(cmd *cobra.Command, args []string) error {
	bucketName := args[0]

	if (!gcpProvider && !awsProvider) || (gcpProvider && awsProvider) {
		return fmt.Errorf("you must specify exactly one provider flag (--gcp or --aws) for the describe command")
	}

	var providerFlag string
	if gcpProvider {
		providerFlag = "gcp"
	} else {
		providerFlag = "aws"
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	storageFormatter := formatter.NewStorageFormatter()
	ctx := cmd.Context()

	client, err := initializeProvider(ctx, providerFlag, cfg)
	if err != nil {
		return fmt.Errorf("error initializing provider: %w", err)
	}
	defer client.Close()

	bucketDetails, err := client.DescribeBucket(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("error describing bucket '%s' on %s: %w", bucketName, providerFlag, err)
	}

	fmt.Println(storageFormatter.FormatBucketDetails(bucketDetails))
	return nil
}
