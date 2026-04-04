package main

import "github.com/spf13/cobra"

func newStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Manage storage resources like buckets and objects",
		Long:  `The storage command allows you to manage storage buckets and objects from configured cloud providers.`,
	}

	cmd.AddCommand(
		newBucketsCmd(),
		newObjectsCmd(),
	)
	return cmd
}
