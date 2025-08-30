// File: cmd/synkronus/storage_cmd.go
package main

import (
	"errors"
	"fmt"
	"strings"
	"synkronus/internal/flags"
	"synkronus/internal/provider/factory"
	"synkronus/internal/provider/registry"

	"github.com/spf13/cobra"
)

// ErrOperationAborted indicates that the user chose not to proceed with a destructive operation
var ErrOperationAborted = errors.New("operation aborted by the user")

type storageFlags struct {
	providersList []string
	provider      string
	location      string
	force         bool
	bucket        string
	prefix        string
}

func newStorageCmd() *cobra.Command {
	cmdFlags := storageFlags{}

	storageCmd := &cobra.Command{
		Use:   "storage",
		Short: "Manage storage resources like buckets and objects",
		Long:  `The storage command allows you to list, describe, create, and delete storage buckets, as well as list and describe objects within them, from configured cloud providers.`,
	}

	// --- Bucket Level Commands ---

	listBucketsCmd := &cobra.Command{
		Use:     "list-buckets",
		Aliases: []string{"list"},
		Short:   "List storage buckets",
		Long: `Lists all storage buckets. If no flags are provided, it queries all configured providers. 
Use the --providers flag to specify which providers to query (e.g., --providers gcp,aws).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			providersToQuery, err := resolveProvidersForList(cmdFlags.providersList, app.ProviderFactory)
			if err != nil {
				return err
			}

			allBuckets, err := app.StorageService.ListAllBuckets(cmd.Context(), providersToQuery)
			if err != nil {
				return err
			}

			if len(allBuckets) > 0 {
				fmt.Println(app.StorageFormatter.FormatBucketList(allBuckets))
			} else {
				if len(providersToQuery) == 0 {
					fmt.Printf("No providers configured. Use 'synkronus config set'. Supported providers: %s\n", strings.Join(registry.GetSupportedProviders(), ", "))
				} else {
					fmt.Println("No buckets found.")
				}
			}
			return nil
		},
	}
	listBucketsCmd.Flags().StringSliceVarP(&cmdFlags.providersList, flags.Providers, flags.ProvidersShort, []string{}, "Specify providers to query (comma-separated). Defaults to all configured providers.")

	describeBucketCmd := &cobra.Command{
		Use:     "describe-bucket [bucket-name]",
		Aliases: []string{"describe"},
		Short:   "Describe a specific storage bucket",
		Long:    `Provides detailed information about a specific storage bucket. You must specify the bucket name and the --provider flag.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			bucketName := args[0]
			providerName := cmdFlags.provider

			bucketDetails, err := app.StorageService.DescribeBucket(cmd.Context(), bucketName, providerName)
			if err != nil {
				return fmt.Errorf("error describing bucket '%s' on %s: %w", bucketName, providerName, err)
			}

			fmt.Println(app.StorageFormatter.FormatBucketDetails(bucketDetails))
			return nil
		},
	}
	describeBucketCmd.Flags().StringVarP(&cmdFlags.provider, flags.Provider, flags.ProviderShort, "", "The provider where the bucket resides (required)")
	describeBucketCmd.MarkFlagRequired(flags.Provider)

	createBucketCmd := &cobra.Command{
		Use:     "create-bucket [bucket-name]",
		Aliases: []string{"create"},
		Short:   "Create a new storage bucket",
		Long:    `Creates a new storage bucket on the specified provider. You must specify the bucket name, the --provider flag, and the --location flag.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			bucketName := args[0]
			providerName := cmdFlags.provider
			err = app.StorageService.CreateBucket(cmd.Context(), bucketName, providerName, cmdFlags.location)
			if err != nil {
				return fmt.Errorf("error creating bucket '%s' on %s: %w", bucketName, providerName, err)
			}

			fmt.Printf("Bucket '%s' created successfully in %s on provider %s.\n", bucketName, cmdFlags.location, providerName)
			return nil
		},
	}
	createBucketCmd.Flags().StringVarP(&cmdFlags.provider, flags.Provider, flags.ProviderShort, "", "The provider to create the bucket on (required)")
	createBucketCmd.MarkFlagRequired(flags.Provider)
	createBucketCmd.Flags().StringVarP(&cmdFlags.location, flags.Location, flags.LocationShort, "", "The location/region to create the bucket in (required)")
	createBucketCmd.MarkFlagRequired(flags.Location)

	deleteBucketCmd := &cobra.Command{
		Use:     "delete-bucket [bucket-name]",
		Aliases: []string{"delete"},
		Short:   "Delete a storage bucket",
		Long: `Deletes a storage bucket on the specified provider. This operation is destructive. 
Confirmation is required by typing the bucket name, unless the --force flag is used.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			bucketName := args[0]
			providerName := cmdFlags.provider

			if !cmdFlags.force {
				warningMessage := fmt.Sprintf("\nWARNING: You are about to delete the bucket '%s' on provider '%s'.\nThis action CANNOT be undone and may result in permanent data loss.", bucketName, strings.ToUpper(providerName))

				confirmed, err := app.Prompter.Confirm(warningMessage, bucketName)
				if err != nil {
					return fmt.Errorf("failed to read confirmation input: %w", err)
				}
				if !confirmed {
					fmt.Println("Deletion aborted: Confirmation mismatch or cancelled.")
					return ErrOperationAborted
				}
			}

			err = app.StorageService.DeleteBucket(cmd.Context(), bucketName, providerName)
			if err != nil {
				return fmt.Errorf("error deleting bucket '%s' on %s: %w", bucketName, providerName, err)
			}

			fmt.Printf("Bucket '%s' deleted successfully from provider %s.\n", bucketName, providerName)
			return nil
		},
	}
	deleteBucketCmd.Flags().StringVarP(&cmdFlags.provider, flags.Provider, flags.ProviderShort, "", "The provider where the bucket resides (required)")
	deleteBucketCmd.MarkFlagRequired(flags.Provider)
	deleteBucketCmd.Flags().BoolVarP(&cmdFlags.force, flags.Force, flags.ForceShort, false, "If set, bypass the interactive confirmation prompt and proceed with deletion")

	// --- Object Level Commands ---

	listObjectsCmd := &cobra.Command{
		Use:   "list-objects",
		Short: "List objects within a storage bucket",
		Long: `Lists objects (files) and common prefixes (directories) within a specified bucket.
Requires the --bucket and --provider flags. Use --prefix to filter the results (e.g., list contents of a specific directory).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			bucketName := cmdFlags.bucket
			providerName := cmdFlags.provider
			prefix := cmdFlags.prefix

			objectList, err := app.StorageService.ListObjects(cmd.Context(), bucketName, providerName, prefix)
			if err != nil {
				return fmt.Errorf("error listing objects in bucket '%s' on %s: %w", bucketName, providerName, err)
			}

			fmt.Println(app.StorageFormatter.FormatObjectList(objectList))
			return nil
		},
	}
	listObjectsCmd.Flags().StringVarP(&cmdFlags.provider, flags.Provider, flags.ProviderShort, "", "The provider where the bucket resides (required)")
	listObjectsCmd.MarkFlagRequired(flags.Provider)
	listObjectsCmd.Flags().StringVarP(&cmdFlags.bucket, flags.Bucket, flags.BucketShort, "", "The name of the bucket to list objects from (required)")
	listObjectsCmd.MarkFlagRequired(flags.Bucket)
	listObjectsCmd.Flags().StringVar(&cmdFlags.prefix, flags.Prefix, "", "Filter results to objects beginning with this prefix (optional)")

	describeObjectCmd := &cobra.Command{
		Use:   "describe-object [object-key]",
		Short: "Describe a specific storage object",
		Long:  `Provides detailed metadata about a specific object within a bucket. Requires the object key as an argument, and the --bucket and --provider flags.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			objectKey := args[0]
			bucketName := cmdFlags.bucket
			providerName := cmdFlags.provider

			objectDetails, err := app.StorageService.DescribeObject(cmd.Context(), bucketName, objectKey, providerName)
			if err != nil {
				return fmt.Errorf("error describing object '%s' in bucket '%s' on %s: %w", objectKey, bucketName, providerName, err)
			}

			fmt.Println(app.StorageFormatter.FormatObjectDetails(objectDetails))
			return nil
		},
	}
	describeObjectCmd.Flags().StringVarP(&cmdFlags.provider, flags.Provider, flags.ProviderShort, "", "The provider where the object resides (required)")
	describeObjectCmd.MarkFlagRequired(flags.Provider)
	describeObjectCmd.Flags().StringVarP(&cmdFlags.bucket, flags.Bucket, flags.BucketShort, "", "The name of the bucket containing the object (required)")
	describeObjectCmd.MarkFlagRequired(flags.Bucket)

	storageCmd.AddCommand(
		listBucketsCmd,
		describeBucketCmd,
		createBucketCmd,
		deleteBucketCmd,
		listObjectsCmd,
		describeObjectCmd,
	)
	return storageCmd
}

func resolveProvidersForList(requestedProviders []string, factory *factory.Factory) ([]string, error) {
	if len(requestedProviders) == 0 {
		return factory.GetConfiguredProviders(), nil
	}

	var validatedProviders []string
	var invalidProviders []string
	seen := make(map[string]bool)

	for _, p := range requestedProviders {
		p = strings.ToLower(strings.TrimSpace(p))

		if seen[p] {
			continue
		}
		seen[p] = true

		if registry.IsSupported(p) {
			if factory.IsConfigured(p) {
				validatedProviders = append(validatedProviders, p)
			} else {
				return nil, fmt.Errorf("provider '%s' was requested but is not configured. Use 'synkronus config set %s.<key> <value>'", p, p)
			}
		} else {
			invalidProviders = append(invalidProviders, p)
		}
	}

	if len(invalidProviders) > 0 {
		return nil, fmt.Errorf("unsupported providers requested: %v. Supported providers are: %v", invalidProviders, registry.GetSupportedProviders())
	}

	return validatedProviders, nil
}
