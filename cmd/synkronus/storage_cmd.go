package main

import "github.com/spf13/cobra"

func newStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Manage storage resources like buckets and objects",
		Long:  `The storage command allows you to list, describe, create, and delete storage buckets, as well as list and describe objects within them, from configured cloud providers.`,
	}

	cmd.AddCommand(
		newListBucketsCmd(),
		newDescribeBucketCmd(),
		newCreateBucketCmd(),
		newDeleteBucketCmd(),
		newListObjectsCmd(),
		newDescribeObjectCmd(),
	)
	return cmd
}
