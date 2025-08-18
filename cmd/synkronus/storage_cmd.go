// File: cmd/synkronus/storage_cmd.go
package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"synkronus/internal/config"
	"synkronus/internal/provider"
	"synkronus/internal/service"
	"synkronus/pkg/formatter"
)

var (
	gcpProvider bool
	awsProvider bool
	location    string
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage storage resources like buckets",
	Long:  `The storage command allows you to list, describe, create, and delete storage buckets from configured cloud providers like AWS and GCP.`,
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

var storageCreateCmd = &cobra.Command{
	Use:   "create [bucket-name]",
	Short: "Create a new storage bucket",
	Long:  `Creates a new storage bucket on the specified provider. You must specify the bucket name, a provider flag (--gcp or --aws), and the location/region flag (--location).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runStorageCreate,
}

var storageDeleteCmd = &cobra.Command{
	Use:   "delete [bucket-name]",
	Short: "Delete a storage bucket",
	Long:  `Deletes a storage bucket on the specified provider. You must specify the bucket name and a provider flag (--gcp or --aws).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runStorageDelete,
}

func init() {
	storageCmd.AddCommand(storageListCmd)
	storageCmd.AddCommand(storageDescribeCmd)
	storageCmd.AddCommand(storageCreateCmd)
	storageCmd.AddCommand(storageDeleteCmd)

	storageCmd.PersistentFlags().BoolVar(&gcpProvider, "gcp", false, "Use GCP provider")
	storageCmd.PersistentFlags().BoolVar(&awsProvider, "aws", false, "Use AWS provider")

	storageCreateCmd.Flags().StringVar(&location, "location", "", "The location/region to create the bucket in (required)")
	storageCreateCmd.MarkFlagRequired("location")
}

func runStorageList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	providerFactory := provider.NewFactory(cfg)
	storageService := service.NewStorageService(providerFactory)
	storageFormatter := formatter.NewStorageFormatter()

	allBuckets, err := storageService.ListAllBuckets(cmd.Context(), gcpProvider, awsProvider)
	if err != nil {
		return err
	}

	if len(allBuckets) > 0 {
		fmt.Println(storageFormatter.FormatBucketList(allBuckets))
	} else {
		fmt.Println("No providers configured or specified, or no buckets found. Configure providers using 'synkronus config set'.")
	}

	return nil
}

func runStorageDescribe(cmd *cobra.Command, args []string) error {
	bucketName := args[0]

	if (!gcpProvider && !awsProvider) || (gcpProvider && awsProvider) {
		return fmt.Errorf("you must specify exactly one provider flag (--gcp or --aws) for the describe command")
	}

	var providerName string
	if gcpProvider {
		providerName = "gcp"
	} else {
		providerName = "aws"
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	providerFactory := provider.NewFactory(cfg)
	storageService := service.NewStorageService(providerFactory)
	storageFormatter := formatter.NewStorageFormatter()

	bucketDetails, err := storageService.DescribeBucket(cmd.Context(), bucketName, providerName)
	if err != nil {
		return fmt.Errorf("error describing bucket '%s' on %s: %w", bucketName, providerName, err)
	}

	fmt.Println(storageFormatter.FormatBucketDetails(bucketDetails))
	return nil
}

func runStorageCreate(cmd *cobra.Command, args []string) error {
	bucketName := args[0]

	if (!gcpProvider && !awsProvider) || (gcpProvider && awsProvider) {
		return fmt.Errorf("you must specify exactly one provider flag (--gcp or --aws) for the create command")
	}

	var providerName string
	if gcpProvider {
		providerName = "gcp"
	} else {
		providerName = "aws"
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	providerFactory := provider.NewFactory(cfg)
	storageService := service.NewStorageService(providerFactory)

	err = storageService.CreateBucket(cmd.Context(), bucketName, providerName, location)
	if err != nil {
		return fmt.Errorf("error creating bucket '%s' on %s: %w", bucketName, providerName, err)
	}

	fmt.Printf("Bucket '%s' created successfully in %s on provider %s.\n", bucketName, location, providerName)
	return nil
}

func runStorageDelete(cmd *cobra.Command, args []string) error {
	bucketName := args[0]

	if (!gcpProvider && !awsProvider) || (gcpProvider && awsProvider) {
		return fmt.Errorf("you must specify exactly one provider flag (--gcp or --aws) for the delete command")
	}

	var providerName string
	if gcpProvider {
		providerName = "gcp"
	} else {
		providerName = "aws"
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	providerFactory := provider.NewFactory(cfg)
	storageService := service.NewStorageService(providerFactory)

	err = storageService.DeleteBucket(cmd.Context(), bucketName, providerName)
	if err != nil {
		return fmt.Errorf("error deleting bucket '%s' on %s: %w", bucketName, providerName, err)
	}

	fmt.Printf("Bucket '%s' deleted successfully from provider %s.\n", bucketName, providerName)
	return nil
}
