package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/spf13/cobra"
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

func getConfigAsMap() (map[string]string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %w", err)
	}

	configMap := make(map[string]string)
	for key, value := range cfg {
		if strValue, ok := value.(string); ok {
			configMap[key] = strValue
		}
	}
	return configMap, nil
}

func runStorageList(cmd *cobra.Command, args []string) error {
	configMap, err := getConfigAsMap()
	if err != nil {
		return err
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
		if val, ok := configMap["gcp_project"]; (ok && val != "" && noFlags) || gcpProvider {
			providersToQuery = append(providersToQuery, "gcp")
		}
		if val, ok := configMap["aws_region"]; (ok && val != "" && noFlags) || awsProvider {
			providersToQuery = append(providersToQuery, "aws")
		}
	}

	if len(providersToQuery) == 0 {
		fmt.Println("No providers configured or specified. Configure GCP/AWS using 'synkronus config set'.")
		return nil
	}

	var allBuckets []storage.Bucket
	var wg sync.WaitGroup
	ctx := context.Background()

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

	configMap, err := getConfigAsMap()
	if err != nil {
		return err
	}

	storageFormatter := formatter.NewStorageFormatter()
	ctx := context.Background()

	client, err := initializeProvider(ctx, providerFlag, configMap)
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
