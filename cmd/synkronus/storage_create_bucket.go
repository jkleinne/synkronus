package main

import (
	"fmt"
	"synkronus/internal/flags"

	"github.com/spf13/cobra"
)

func newCreateBucketCmd() *cobra.Command {
	var provider string
	var location string

	cmd := &cobra.Command{
		Use:     "create-bucket [bucket-name]",
		Aliases: []string{"create"},
		Short:   "Create a new storage bucket",
		Long:    `Creates a new storage bucket on the specified provider. You must specify the bucket name, the --provider flag, and the --location flag.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			bucketName := args[0]
			err = app.StorageService.CreateBucket(cmd.Context(), bucketName, provider, location)
			if err != nil {
				return fmt.Errorf("error creating bucket '%s' on %s: %w", bucketName, provider, err)
			}

			fmt.Printf("Bucket '%s' created successfully in %s on provider %s.\n", bucketName, location, provider)
			return nil
		},
	}
	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider to create the bucket on (required)")
	cmd.MarkFlagRequired(flags.Provider)
	cmd.Flags().StringVarP(&location, flags.Location, flags.LocationShort, "", "The location/region to create the bucket in (required)")
	cmd.MarkFlagRequired(flags.Location)

	return cmd
}
