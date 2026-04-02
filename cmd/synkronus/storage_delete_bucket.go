package main

import (
	"fmt"
	"strings"
	"synkronus/internal/flags"

	"github.com/spf13/cobra"
)

func newDeleteBucketCmd() *cobra.Command {
	var provider string
	var force bool

	cmd := &cobra.Command{
		Use:     "delete-bucket [bucket-name]",
		Aliases: []string{"delete"},
		Short:   "Delete a storage bucket",
		Long: `Deletes a storage bucket on the specified provider. This operation is destructive.
Confirmation is required by typing the bucket name, unless the --force flag is used.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			bucketName := args[0]

			if !force {
				warningMessage := fmt.Sprintf("\nWARNING: You are about to delete the bucket '%s' on provider '%s'.\nThis action CANNOT be undone and may result in permanent data loss.", bucketName, strings.ToUpper(provider))

				confirmed, err := app.Prompter.Confirm(warningMessage, bucketName)
				if err != nil {
					return fmt.Errorf("failed to read confirmation input: %w", err)
				}
				if !confirmed {
					fmt.Println("Deletion aborted: Confirmation mismatch or cancelled.")
					return ErrOperationAborted
				}
			}

			err = app.StorageService.DeleteBucket(cmd.Context(), bucketName, provider)
			if err != nil {
				return fmt.Errorf("error deleting bucket '%s' on %s: %w", bucketName, provider, err)
			}

			fmt.Printf("Bucket '%s' deleted successfully from provider %s.\n", bucketName, provider)
			return nil
		},
	}
	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider where the bucket resides (required)")
	cmd.MarkFlagRequired(flags.Provider)
	cmd.Flags().BoolVarP(&force, flags.Force, flags.ForceShort, false, "If set, bypass the interactive confirmation prompt and proceed with deletion")

	return cmd
}
