package main

import "github.com/spf13/cobra"

func newObjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "objects",
		Short: "Manage storage objects",
		Long:  `List, describe, download, and manage objects within storage buckets.`,
	}

	cmd.AddCommand(
		newListObjectsCmd(),
		newDescribeObjectCmd(),
		newDownloadObjectCmd(),
		newUploadObjectCmd(),
		newDeleteObjectCmd(),
		newCopyObjectCmd(),
	)
	return cmd
}
