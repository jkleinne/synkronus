package main

import (
	"os"
	"synkronus/internal/flags"
	"synkronus/internal/output"

	"github.com/spf13/cobra"
)

func newListObjectsCmd() *cobra.Command {
	var provider string
	var bucket string
	var prefix string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List objects within a storage bucket",
		Long: `Lists objects (files) and common prefixes (directories) within a specified bucket.
Requires the --bucket and --provider flags. Use --prefix to filter the results (e.g., list contents of a specific directory).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			objectList, err := app.StorageService.ListObjects(cmd.Context(), bucket, provider, prefix)
			if err != nil {
				return err
			}

			return output.Render(os.Stdout, app.OutputFormat, output.ObjectListView{ObjectList: objectList})
		},
	}
	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider where the bucket resides (required)")
	cmd.MarkFlagRequired(flags.Provider)
	cmd.Flags().StringVarP(&bucket, flags.Bucket, flags.BucketShort, "", "The name of the bucket to list objects from (required)")
	cmd.MarkFlagRequired(flags.Bucket)
	cmd.Flags().StringVar(&prefix, flags.Prefix, "", "Filter results to objects beginning with this prefix (optional)")

	return cmd
}
