package main

import (
	"fmt"
	"strings"
	"synkronus/internal/flags"

	"github.com/spf13/cobra"
)

func newDeleteObjectCmd() *cobra.Command {
	var provider string
	var bucket string
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [object-key]",
		Short: "Delete a storage object",
		Long: `Deletes an object from a storage bucket. This operation is destructive.
Confirmation is required by typing the object key, unless the --force flag is used.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			objectKey := args[0]

			if !force {
				warningMessage := fmt.Sprintf(
					"\nWARNING: You are about to delete object '%s' from bucket '%s' (%s).\nThis action cannot be undone.",
					objectKey, bucket, strings.ToUpper(provider))

				confirmed, err := app.Prompter.Confirm(warningMessage, objectKey)
				if err != nil {
					return fmt.Errorf("failed to read confirmation input: %w", err)
				}
				if !confirmed {
					fmt.Println("Deletion aborted: Confirmation mismatch or cancelled.")
					return ErrOperationAborted
				}
			}

			if err := app.StorageService.DeleteObject(cmd.Context(), bucket, objectKey, provider); err != nil {
				return fmt.Errorf("error deleting object '%s' from bucket '%s' on %s: %w", objectKey, bucket, provider, err)
			}

			fmt.Printf("Object '%s' deleted successfully from bucket '%s' on provider %s.\n", objectKey, bucket, provider)
			return nil
		},
	}

	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider where the object resides (required)")
	cmd.MarkFlagRequired(flags.Provider)
	cmd.Flags().StringVarP(&bucket, flags.Bucket, flags.BucketShort, "", "The name of the bucket containing the object (required)")
	cmd.MarkFlagRequired(flags.Bucket)
	cmd.Flags().BoolVarP(&force, flags.Force, flags.ForceShort, false, "Bypass interactive confirmation prompt")

	return cmd
}
