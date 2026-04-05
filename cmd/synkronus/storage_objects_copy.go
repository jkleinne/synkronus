package main

import (
	"fmt"
	"synkronus/internal/flags"

	"github.com/spf13/cobra"
)

func newCopyObjectCmd() *cobra.Command {
	var provider string
	var bucket string
	var destBucket string
	var destKey string

	cmd := &cobra.Command{
		Use:   "copy [src-key]",
		Short: "Copy a storage object",
		Long: `Copies an object within the same provider. The --bucket flag specifies the source bucket.
If --dest-key is omitted, the source key is reused. Same-bucket copy is supported (e.g., rename by copying to a new key).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			srcKey := args[0]

			if destKey == "" {
				destKey = srcKey
			}

			if err := app.StorageService.CopyObject(cmd.Context(), bucket, srcKey, destBucket, destKey, provider); err != nil {
				return err
			}

			fmt.Printf("Object '%s' copied successfully from bucket '%s' to '%s/%s' on provider %s.\n",
				srcKey, bucket, destBucket, destKey, provider)
			return nil
		},
	}

	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider for the copy operation (required)")
	cmd.MarkFlagRequired(flags.Provider)
	cmd.Flags().StringVarP(&bucket, flags.Bucket, flags.BucketShort, "", "The source bucket (required)")
	cmd.MarkFlagRequired(flags.Bucket)
	cmd.Flags().StringVar(&destBucket, flags.DestBucket, "", "The destination bucket (required)")
	cmd.MarkFlagRequired(flags.DestBucket)
	cmd.Flags().StringVar(&destKey, flags.DestKey, "", "Destination object key (defaults to source key)")

	return cmd
}
