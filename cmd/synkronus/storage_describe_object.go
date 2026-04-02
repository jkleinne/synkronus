package main

import (
	"fmt"
	"os"
	"synkronus/internal/flags"
	"synkronus/internal/output"

	"github.com/spf13/cobra"
)

func newDescribeObjectCmd() *cobra.Command {
	var provider string
	var bucket string

	cmd := &cobra.Command{
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

			objectDetails, err := app.StorageService.DescribeObject(cmd.Context(), bucket, objectKey, provider)
			if err != nil {
				return fmt.Errorf("error describing object '%s' in bucket '%s' on %s: %w", objectKey, bucket, provider, err)
			}

			return output.Render(os.Stdout, app.OutputFormat, output.ObjectDetailView{Object: objectDetails})
		},
	}
	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider where the object resides (required)")
	cmd.MarkFlagRequired(flags.Provider)
	cmd.Flags().StringVarP(&bucket, flags.Bucket, flags.BucketShort, "", "The name of the bucket containing the object (required)")
	cmd.MarkFlagRequired(flags.Bucket)

	return cmd
}
