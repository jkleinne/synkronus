// File: cmd/synkronus/storage_cmd.go
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStorageCmd(app *appContainer) *cobra.Command {
	var (
		gcpProvider bool
		awsProvider bool
		location    string
	)

	storageCmd := &cobra.Command{
		Use:   "storage",
		Short: "Manage storage resources like buckets",
		Long:  `The storage command allows you to list, describe, create, and delete storage buckets from configured cloud providers like AWS and GCP.`,
	}

	storageCmd.PersistentFlags().BoolVar(&gcpProvider, "gcp", false, "Use GCP provider")
	storageCmd.PersistentFlags().BoolVar(&awsProvider, "aws", false, "Use AWS provider")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List storage buckets",
		Long:  `Lists all storage buckets from the configured cloud providers. Use flags to specify a provider.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			allBuckets, err := app.StorageService.ListAllBuckets(cmd.Context(), gcpProvider, awsProvider)
			if err != nil {
				return err
			}

			if len(allBuckets) > 0 {
				fmt.Println(app.StorageFormatter.FormatBucketList(allBuckets))
			} else {
				fmt.Println("No providers configured or specified, or no buckets found. Configure providers using 'synkronus config set'.")
			}
			return nil
		},
	}

	describeCmd := &cobra.Command{
		Use:   "describe [bucket-name]",
		Short: "Describe a specific storage bucket",
		Long:  `Provides detailed information about a specific storage bucket. You must specify the bucket name and the provider flag (--gcp or --aws).`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			bucketDetails, err := app.StorageService.DescribeBucket(cmd.Context(), bucketName, providerName)
			if err != nil {
				return fmt.Errorf("error describing bucket '%s' on %s: %w", bucketName, providerName, err)
			}

			fmt.Println(app.StorageFormatter.FormatBucketDetails(bucketDetails))
			return nil
		},
	}

	createCmd := &cobra.Command{
		Use:   "create [bucket-name]",
		Short: "Create a new storage bucket",
		Long:  `Creates a new storage bucket on the specified provider. You must specify the bucket name, a provider flag (--gcp or --aws), and the location/region flag (--location).`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			err := app.StorageService.CreateBucket(cmd.Context(), bucketName, providerName, location)
			if err != nil {
				return fmt.Errorf("error creating bucket '%s' on %s: %w", bucketName, providerName, err)
			}

			fmt.Printf("Bucket '%s' created successfully in %s on provider %s.\n", bucketName, location, providerName)
			return nil
		},
	}
	createCmd.Flags().StringVar(&location, "location", "", "The location/region to create the bucket in (required)")
	createCmd.MarkFlagRequired("location")

	deleteCmd := &cobra.Command{
		Use:   "delete [bucket-name]",
		Short: "Delete a storage bucket",
		Long:  `Deletes a storage bucket on the specified provider. You must specify the bucket name and a provider flag (--gcp or --aws).`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			err := app.StorageService.DeleteBucket(cmd.Context(), bucketName, providerName)
			if err != nil {
				return fmt.Errorf("error deleting bucket '%s' on %s: %w", bucketName, providerName, err)
			}

			fmt.Printf("Bucket '%s' deleted successfully from provider %s.\n", bucketName, providerName)
			return nil
		},
	}

	storageCmd.AddCommand(listCmd, describeCmd, createCmd, deleteCmd)
	return storageCmd
}
