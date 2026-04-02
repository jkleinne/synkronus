package main

import (
	"fmt"
	"os"
	"synkronus/internal/flags"
	"synkronus/internal/output"

	"github.com/spf13/cobra"
)

func newDescribeInstanceCmd() *cobra.Command {
	var provider string

	cmd := &cobra.Command{
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

			instanceDetails, err := app.SqlService.DescribeInstance(cmd.Context(), instanceName, provider)
			if err != nil {
				return fmt.Errorf("error describing SQL instance '%s' on %s: %w", instanceName, provider, err)
			}

			return output.Render(os.Stdout, app.OutputFormat, output.InstanceDetailView{Instance: instanceDetails})
		},
	}
	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider where the SQL instance resides (required)")
	cmd.MarkFlagRequired(flags.Provider)

	return cmd
}
