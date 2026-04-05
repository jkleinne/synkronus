package main

import (
	"os"
	"synkronus/internal/flags"
	"synkronus/internal/output"

	"github.com/spf13/cobra"
)

func newDescribeBucketCmd() *cobra.Command {
	var provider string

	cmd := &cobra.Command{
		Use:   "describe [bucket-name]",
		Short: "Describe a specific storage bucket",
		Long:    `Provides detailed information about a specific storage bucket. You must specify the bucket name and the --provider flag.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			bucketName := args[0]

			bucketDetails, err := app.StorageService.DescribeBucket(cmd.Context(), bucketName, provider)
			if err != nil {
				return err
			}

			return output.Render(os.Stdout, app.OutputFormat, output.BucketDetailView{Bucket: bucketDetails})
		},
	}
	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider where the bucket resides (required)")
	cmd.MarkFlagRequired(flags.Provider)

	return cmd
}
