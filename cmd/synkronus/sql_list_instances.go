package main

import (
	"fmt"
	"os"
	"strings"
	"synkronus/internal/flags"
	"synkronus/internal/output"

	"github.com/spf13/cobra"
)

func newListInstancesCmd() *cobra.Command {
	var providersList []string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"list-instances"},
		Short:   "List SQL database instances",
		Long: `Lists all SQL database instances. If no flags are provided, it queries all configured SQL providers.
Use the --providers flag to specify which providers to query (e.g., --providers gcp).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			resolver := &ProviderResolver{
				IsSupported:   isInList(app.ProviderFactory.SupportedSqlProviders),
				IsConfigured:  app.ProviderFactory.IsSqlConfigured,
				GetConfigured: app.ProviderFactory.ConfiguredSqlProviders,
				GetSupported:  app.ProviderFactory.SupportedSqlProviders,
				Label:         "SQL",
			}
			providersToQuery, err := resolver.Resolve(providersList)
			if err != nil {
				return err
			}

			allInstances, err := app.SqlService.ListAllInstances(cmd.Context(), providersToQuery)
			if err != nil && len(allInstances) == 0 {
				return err
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: some SQL providers failed: %v\n", err)
			}

			if len(allInstances) == 0 {
				if len(providersToQuery) == 0 {
					fmt.Printf("No SQL providers configured. Use 'synkronus config set'. Supported SQL providers: %s\n", strings.Join(app.ProviderFactory.SupportedSqlProviders(), ", "))
				} else {
					fmt.Println("No SQL instances found.")
				}
				return nil
			}
			return output.Render(os.Stdout, app.OutputFormat, output.InstanceListView(allInstances))
		},
	}
	cmd.Flags().StringSliceVarP(&providersList, flags.Providers, flags.ProvidersShort, nil, "Specify providers to query (comma-separated). Defaults to all configured SQL providers.")

	return cmd
}
