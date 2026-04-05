package main

import (
	"fmt"
	"os"
	"synkronus/internal/domain/storage"
	"synkronus/internal/flags"
	"synkronus/internal/output"

	"github.com/spf13/cobra"
)

func newListObjectsCmd() *cobra.Command {
	var provider string
	var bucket string
	var prefix string
	var maxResults int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List objects within a storage bucket",
		Long: `Lists objects (files) and common prefixes (directories) within a specified bucket.
Requires the --bucket and --provider flags. Use --prefix to filter the results (e.g., list contents of a specific directory).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if maxResults < 0 {
				return fmt.Errorf("--%s must be a non-negative integer, got %d", flags.MaxResults, maxResults)
			}

			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			objectList, err := app.StorageService.ListObjects(cmd.Context(), bucket, provider, prefix, maxResults)
			if err != nil {
				return fmt.Errorf("error listing objects in bucket '%s' on %s: %w", bucket, provider, err)
			}

			return output.Render(os.Stdout, app.OutputFormat, output.ObjectListView{ObjectList: objectList})
		},
	}
	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider where the bucket resides (required)")
	cmd.MarkFlagRequired(flags.Provider)
	cmd.Flags().StringVarP(&bucket, flags.Bucket, flags.BucketShort, "", "The name of the bucket to list objects from (required)")
	cmd.MarkFlagRequired(flags.Bucket)
	cmd.Flags().StringVar(&prefix, flags.Prefix, "", "Filter results to objects beginning with this prefix (optional)")
	cmd.Flags().IntVar(&maxResults, flags.MaxResults, storage.DefaultMaxResults, "Maximum number of items to return (0 for unlimited)")

	return cmd
}
