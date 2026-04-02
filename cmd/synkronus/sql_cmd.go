// File: cmd/synkronus/sql_cmd.go
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

type sqlFlags struct {
	providersList []string
	provider      string
}

func newSqlCmd() *cobra.Command {
	cmdFlags := sqlFlags{}

	sqlCmd := &cobra.Command{
		Use:   "sql",
		Short: "Manage SQL resources",
		Long:  `The sql command allows you to interact with SQL database instances from various cloud providers.`,
	}

	// --- Instance Level Commands ---

	listInstancesCmd := &cobra.Command{
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

			providersToQuery, err := resolveProviders(
				cmdFlags.providersList,
				registry.IsSqlSupported,
				app.ProviderFactory.IsSqlConfigured,
				app.ProviderFactory.GetConfiguredSqlProviders,
				registry.GetSupportedSqlProviders,
				"SQL",
			)
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

			if len(allInstances) > 0 {
				return output.Render(os.Stdout, app.OutputFormat, output.InstanceListView(allInstances))
			} else {
				if len(providersToQuery) == 0 {
					fmt.Printf("No SQL providers configured. Use 'synkronus config set'. Supported SQL providers: %s\n", strings.Join(registry.GetSupportedSqlProviders(), ", "))
				} else {
					fmt.Println("No SQL instances found.")
				}
			}
			return nil
		},
	}
	listInstancesCmd.Flags().StringSliceVarP(&cmdFlags.providersList, flags.Providers, flags.ProvidersShort, []string{}, "Specify providers to query (comma-separated). Defaults to all configured SQL providers.")

	describeInstanceCmd := &cobra.Command{
		Use:     "describe [instance-name]",
		Aliases: []string{"describe-instance"},
		Short:   "Describe a specific SQL database instance",
		Long:    `Provides detailed information about a specific SQL database instance. You must specify the instance name and the --provider flag.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			instanceName := args[0]
			providerName := cmdFlags.provider

			instanceDetails, err := app.SqlService.DescribeInstance(cmd.Context(), instanceName, providerName)
			if err != nil {
				return fmt.Errorf("error describing SQL instance '%s' on %s: %w", instanceName, providerName, err)
			}

			return output.Render(os.Stdout, app.OutputFormat, output.InstanceDetailView{Instance: instanceDetails})
		},
	}
	describeInstanceCmd.Flags().StringVarP(&cmdFlags.provider, flags.Provider, flags.ProviderShort, "", "The provider where the SQL instance resides (required)")
	describeInstanceCmd.MarkFlagRequired(flags.Provider)

	sqlCmd.AddCommand(listInstancesCmd, describeInstanceCmd)
	return sqlCmd
}

