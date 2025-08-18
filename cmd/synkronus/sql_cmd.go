// File: cmd/synkronus/sql_cmd.go
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSqlCmd() *cobra.Command {
	sqlCmd := &cobra.Command{
		Use:   "sql",
		Short: "Manage SQL resources",
		Long:  `The sql command allows you to interact with SQL database instances from various cloud providers.`,
	}

	sqlListCmd := &cobra.Command{
		Use:   "list",
		Short: "List SQL resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing SQL resources...")
			// TODO: Implement SQL listing functionality
			return nil
		},
	}

	sqlCmd.AddCommand(sqlListCmd)
	return sqlCmd
}
