package main

import "github.com/spf13/cobra"

func newSqlCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sql",
		Short: "Manage SQL resources",
		Long:  `The sql command allows you to interact with SQL database instances from various cloud providers.`,
	}

	cmd.AddCommand(newListInstancesCmd(), newDescribeInstanceCmd())
	return cmd
}
