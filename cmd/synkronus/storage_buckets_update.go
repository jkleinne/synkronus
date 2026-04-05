package main

import (
	"fmt"
	"strings"
	"synkronus/internal/domain/storage"
	"synkronus/internal/flags"

	"github.com/spf13/cobra"
)

func newUpdateBucketCmd() *cobra.Command {
	var provider string
	var setLabels map[string]string
	var removeLabelsRaw string
	var versioning bool

	cmd := &cobra.Command{
		Use:   "update [bucket-name]",
		Short: "Update properties of an existing storage bucket",
		Long:  `Updates labels and/or versioning on an existing storage bucket. At least one mutation flag must be provided.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			changed := cmd.Flags().Changed(flags.SetLabels) ||
				cmd.Flags().Changed(flags.RemoveLabels) ||
				cmd.Flags().Changed(flags.VersioningFlag)
			if !changed {
				return fmt.Errorf("at least one of --%s, --%s, or --%s must be specified",
					flags.SetLabels, flags.RemoveLabels, flags.VersioningFlag)
			}

			// Validate set-labels keys
			for k := range setLabels {
				if k == "" {
					return fmt.Errorf("invalid --%s: label key cannot be empty", flags.SetLabels)
				}
			}

			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			opts := storage.UpdateBucketOptions{
				Name: args[0],
			}
			if cmd.Flags().Changed(flags.SetLabels) {
				opts.SetLabels = setLabels
			}
			if cmd.Flags().Changed(flags.RemoveLabels) && removeLabelsRaw != "" {
				for _, part := range strings.Split(removeLabelsRaw, ",") {
					trimmed := strings.TrimSpace(part)
					if trimmed == "" {
						return fmt.Errorf("invalid --%s: label key cannot be empty", flags.RemoveLabels)
					}
					opts.RemoveLabels = append(opts.RemoveLabels, trimmed)
				}
			}
			if cmd.Flags().Changed(flags.VersioningFlag) {
				opts.Versioning = &versioning
			}

			err = app.StorageService.UpdateBucket(cmd.Context(), opts, provider)
			if err != nil {
				return fmt.Errorf("error updating bucket '%s' on %s: %w", opts.Name, provider, err)
			}

			fmt.Printf("Bucket '%s' updated successfully on provider %s.\n", opts.Name, provider)
			return nil
		},
	}

	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider of the bucket to update (required)")
	cmd.MarkFlagRequired(flags.Provider)
	cmd.Flags().StringToStringVar(&setLabels, flags.SetLabels, nil, "Labels to add or overwrite (e.g. --set-labels env=prod,team=data)")
	cmd.Flags().StringVar(&removeLabelsRaw, flags.RemoveLabels, "", "Label keys to remove, comma-separated (e.g. --remove-labels old-key,temp)")
	cmd.Flags().BoolVar(&versioning, flags.VersioningFlag, false, "Enable or disable object versioning (--versioning=true or --versioning=false)")

	return cmd
}
