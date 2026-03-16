// File: cmd/synkronus/sql_cmd.go
package main

import (
	"fmt"
	"os"
	"strings"
	"synkronus/internal/flags"
	"synkronus/internal/provider/factory"
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

			providersToQuery, err := resolveSqlProvidersForList(cmdFlags.providersList, app.ProviderFactory)
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
				fmt.Println(app.SqlFormatter.FormatInstanceList(allInstances))
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

			fmt.Println(app.SqlFormatter.FormatInstanceDetails(instanceDetails))
			return nil
		},
	}
	describeInstanceCmd.Flags().StringVarP(&cmdFlags.provider, flags.Provider, flags.ProviderShort, "", "The provider where the SQL instance resides (required)")
	describeInstanceCmd.MarkFlagRequired(flags.Provider)

	sqlCmd.AddCommand(listInstancesCmd, describeInstanceCmd)
	return sqlCmd
}

// resolveSqlProvidersForList validates and resolves the list of SQL providers to query
func resolveSqlProvidersForList(requestedProviders []string, f *factory.Factory) ([]string, error) {
	if len(requestedProviders) == 0 {
		return f.GetConfiguredSqlProviders(), nil
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

		if registry.IsSqlSupported(p) {
			if f.IsSqlConfigured(p) {
				validatedProviders = append(validatedProviders, p)
			} else {
				return nil, fmt.Errorf("SQL provider '%s' was requested but is not configured. Use 'synkronus config set %s.<key> <value>'", p, p)
			}
		} else {
			invalidProviders = append(invalidProviders, p)
		}
	}

	if len(invalidProviders) > 0 {
		return nil, fmt.Errorf("unsupported SQL providers requested: %v. Supported SQL providers are: %v", invalidProviders, registry.GetSupportedSqlProviders())
	}

	return validatedProviders, nil
}
