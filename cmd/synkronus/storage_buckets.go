package main

import "github.com/spf13/cobra"

func newBucketsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buckets",
		Short: "Manage storage buckets",
		Long:  `List, describe, create, and delete storage buckets across configured cloud providers.`,
	}

	cmd.AddCommand(
		newListBucketsCmd(),
		newDescribeBucketCmd(),
		newCreateBucketCmd(),
		newDeleteBucketCmd(),
	)
	return cmd
}
