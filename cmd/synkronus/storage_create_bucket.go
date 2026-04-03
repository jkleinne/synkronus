package main

import (
	"fmt"
	"strings"
	"synkronus/internal/domain/storage"
	"synkronus/internal/flags"

	"github.com/spf13/cobra"
)

func newCreateBucketCmd() *cobra.Command {
	var provider string
	var location string
	var storageClass string
	var labels map[string]string
	var versioning bool
	var uniformAccess bool
	var publicAccessPrevention string

	cmd := &cobra.Command{
		Use:     "create-bucket [bucket-name]",
		Aliases: []string{"create"},
		Short:   "Create a new storage bucket",
		Long:    `Creates a new storage bucket on the specified provider. You must specify the bucket name, the --provider flag, and the --location flag. Optional flags control storage class, labels, versioning, uniform access, and public access prevention.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			opts := storage.CreateBucketOptions{
				Name:     args[0],
				Location: location,
			}
			if storageClass != "" {
				opts.StorageClass = strings.ToUpper(storageClass)
			}
			if len(labels) > 0 {
				opts.Labels = labels
			}
			if cmd.Flags().Changed(flags.VersioningFlag) {
				opts.Versioning = &versioning
			}
			if cmd.Flags().Changed(flags.UniformAccess) {
				if strings.ToLower(provider) != "gcp" {
					return fmt.Errorf("--%s is only supported for GCP", flags.UniformAccess)
				}
				opts.UniformBucketLevelAccess = &uniformAccess
			}
			if cmd.Flags().Changed(flags.PublicAccessPreventionFlag) {
				normalized := strings.ToLower(publicAccessPrevention)
				if normalized != storage.PublicAccessPreventionEnforced && normalized != storage.PublicAccessPreventionInherited {
					return fmt.Errorf("invalid --public-access-prevention value %q: must be \"enforced\" or \"inherited\"", publicAccessPrevention)
				}
				opts.PublicAccessPrevention = &normalized
			}

			err = app.StorageService.CreateBucket(cmd.Context(), opts, provider)
			if err != nil {
				return fmt.Errorf("error creating bucket '%s' on %s: %w", opts.Name, provider, err)
			}

			fmt.Printf("Bucket '%s' created successfully in %s on provider %s.\n", opts.Name, location, provider)
			return nil
		},
	}

	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider to create the bucket on (required)")
	cmd.MarkFlagRequired(flags.Provider)
	cmd.Flags().StringVarP(&location, flags.Location, flags.LocationShort, "", "The location/region to create the bucket in (required)")
	cmd.MarkFlagRequired(flags.Location)

	cmd.Flags().StringVarP(&storageClass, flags.StorageClass, flags.StorageClassShort, "", "Storage class (STANDARD, NEARLINE, COLDLINE, ARCHIVE)")
	cmd.Flags().StringToStringVar(&labels, flags.Labels, nil, "Labels as key=value pairs (e.g. --labels env=prod,team=data)")
	cmd.Flags().BoolVar(&versioning, flags.VersioningFlag, false, "Enable object versioning")
	cmd.Flags().BoolVar(&uniformAccess, flags.UniformAccess, false, "Enable Uniform Bucket-Level Access")
	cmd.Flags().StringVar(&publicAccessPrevention, flags.PublicAccessPreventionFlag, "", "Public access prevention (enforced or inherited)")

	return cmd
}
