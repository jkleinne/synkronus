// File: cmd/synkronus/storage_cmd.go
package main

import (
	"fmt"
	"strings"
	"synkronus/internal/flags"
	"synkronus/internal/provider"

	"github.com/spf13/cobra"
)

type storageFlags struct {
	providersList []string
	provider      string
	location      string
}

func newStorageCmd(app *appContainer) *cobra.Command {
	cmdFlags := storageFlags{}

	storageCmd := &cobra.Command{
		Use:   "storage",
		Short: "Manage storage resources like buckets",
		Long:  `The storage command allows you to list, describe, create, and delete storage buckets from configured cloud providers.`,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List storage buckets",
		Long: `Lists all storage buckets. If no flags are provided, it queries all configured providers. 
Use the --providers flag to specify which providers to query (e.g., --providers gcp,aws).`,
		RunE: func(cmd *cobra.Command, args []string) error {
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
					fmt.Printf("No providers configured. Use 'synkronus config set'. Supported providers: %s\n", strings.Join(provider.GetSupportedProviders(), ", "))
				} else {
					fmt.Println("No buckets found.")
				}
			}
			return nil
		},
	}
	listCmd.Flags().StringSliceVarP(&cmdFlags.providersList, flags.Providers, flags.ProvidersShort, []string{}, "Specify providers to query (comma-separated). Defaults to all configured providers.")

	describeCmd := &cobra.Command{
		Use:   "describe [bucket-name]",
		Short: "Describe a specific storage bucket",
		Long:  `Provides detailed information about a specific storage bucket. You must specify the bucket name and the --provider flag.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
	describeCmd.Flags().StringVarP(&cmdFlags.provider, flags.Provider, flags.ProviderShort, "", "The provider where the bucket resides (required)")
	describeCmd.MarkFlagRequired(flags.Provider)

	createCmd := &cobra.Command{
		Use:   "create [bucket-name]",
		Short: "Create a new storage bucket",
		Long:  `Creates a new storage bucket on the specified provider. You must specify the bucket name, the --provider flag, and the --location flag.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bucketName := args[0]
			providerName := cmdFlags.provider
			err := app.StorageService.CreateBucket(cmd.Context(), bucketName, providerName, cmdFlags.location)
			if err != nil {
				return fmt.Errorf("error creating bucket '%s' on %s: %w", bucketName, providerName, err)
			}

			fmt.Printf("Bucket '%s' created successfully in %s on provider %s.\n", bucketName, cmdFlags.location, providerName)
			return nil
		},
	}
	createCmd.Flags().StringVarP(&cmdFlags.provider, flags.Provider, flags.ProviderShort, "", "The provider to create the bucket on (required)")
	createCmd.MarkFlagRequired(flags.Provider)
	createCmd.Flags().StringVarP(&cmdFlags.location, flags.Location, flags.LocationShort, "", "The location/region to create the bucket in (required)")
	createCmd.MarkFlagRequired(flags.Location)

	deleteCmd := &cobra.Command{
		Use:   "delete [bucket-name]",
		Short: "Delete a storage bucket",
		Long:  `Deletes a storage bucket on the specified provider. You must specify the bucket name and the --provider flag.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bucketName := args[0]
			providerName := cmdFlags.provider

			err := app.StorageService.DeleteBucket(cmd.Context(), bucketName, providerName)
			if err != nil {
				return fmt.Errorf("error deleting bucket '%s' on %s: %w", bucketName, providerName, err)
			}

			fmt.Printf("Bucket '%s' deleted successfully from provider %s.\n", bucketName, providerName)
			return nil
		},
	}
	deleteCmd.Flags().StringVarP(&cmdFlags.provider, flags.Provider, flags.ProviderShort, "", "The provider where the bucket resides (required)")
	deleteCmd.MarkFlagRequired(flags.Provider)

	storageCmd.AddCommand(listCmd, describeCmd, createCmd, deleteCmd)
	return storageCmd
}

func resolveProvidersForList(requestedProviders []string, factory *provider.Factory) ([]string, error) {
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

		if provider.IsSupported(p) {
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
		return nil, fmt.Errorf("unsupported providers requested: %v. Supported providers are: %v", invalidProviders, provider.GetSupportedProviders())
	}

	return validatedProviders, nil
}
