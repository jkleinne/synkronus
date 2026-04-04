package main

import (
	"fmt"
	"os"
	"strings"
	"synkronus/internal/flags"
	"synkronus/internal/output"
	"synkronus/internal/provider/registry"

	"github.com/spf13/cobra"
)

func newListBucketsCmd() *cobra.Command {
	var providersList []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List storage buckets",
		Long: `Lists all storage buckets. If no flags are provided, it queries all configured providers.
Use the --providers flag to specify which providers to query (e.g., --providers gcp,aws).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			resolver := &ProviderResolver{
				IsSupported:   registry.IsSupported,
				IsConfigured:  app.ProviderFactory.IsConfigured,
				GetConfigured: app.ProviderFactory.GetConfiguredProviders,
				GetSupported:  registry.GetSupportedProviders,
				Label:         "storage",
			}
			providersToQuery, err := resolver.Resolve(providersList)
			if err != nil {
				return err
			}

			allBuckets, err := app.StorageService.ListAllBuckets(cmd.Context(), providersToQuery)
			if err != nil && len(allBuckets) == 0 {
				return err
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: some providers failed: %v\n", err)
			}

			if len(allBuckets) > 0 {
				return output.Render(os.Stdout, app.OutputFormat, output.BucketListView(allBuckets))
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
	cmd.Flags().StringSliceVarP(&providersList, flags.Providers, flags.ProvidersShort, nil, "Specify providers to query (comma-separated). Defaults to all configured providers.")

	return cmd
}
