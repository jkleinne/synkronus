package main

import (
	"fmt"
	"os"
	"strings"
	"synkronus/internal/domain/storage"
	"synkronus/internal/flags"
	"synkronus/internal/provider/storage/shared"

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
		Use:   "create [bucket-name]",
		Short: "Create a new storage bucket",
		Long:    `Creates a new storage bucket on the specified provider. You must specify the bucket name, the --provider flag, and the --location flag. Optional flags control storage class, labels, versioning, uniform access, and public access prevention.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			opts, err := buildCreateBucketOptions(cmd, args[0], provider, location, storageClass, publicAccessPrevention, labels, versioning, uniformAccess)
			if err != nil {
				return err
			}

			result, err := app.StorageService.CreateBucket(cmd.Context(), opts, provider)
			if err != nil {
				return err
			}

			fmt.Printf("Bucket '%s' created successfully in %s on provider %s.\n", opts.Name, location, provider)
			for _, w := range result.Warnings {
				fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
			}
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

// buildCreateBucketOptions assembles a CreateBucketOptions from the parsed flag values.
// Flag-changed checks are used rather than zero-value checks so that explicitly passing
// false (e.g. --versioning=false) is still honored.
func buildCreateBucketOptions(
	cmd *cobra.Command,
	name, provider, location, storageClass, publicAccessPrevention string,
	labels map[string]string,
	versioning, uniformAccess bool,
) (storage.CreateBucketOptions, error) {
	opts := storage.CreateBucketOptions{
		Name:     name,
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
		if !shared.SupportsOption(provider, "uniform-access") {
			return storage.CreateBucketOptions{}, fmt.Errorf("--%s is not supported for provider %q", flags.UniformAccess, provider)
		}
		opts.UniformBucketLevelAccess = &uniformAccess
	}
	if cmd.Flags().Changed(flags.PublicAccessPreventionFlag) {
		normalized := strings.ToLower(publicAccessPrevention)
		if normalized != storage.PublicAccessPreventionEnforced && normalized != storage.PublicAccessPreventionInherited {
			return storage.CreateBucketOptions{}, fmt.Errorf("invalid --public-access-prevention value %q: must be \"enforced\" or \"inherited\"", publicAccessPrevention)
		}
		opts.PublicAccessPrevention = &normalized
	}

	return opts, nil
}
