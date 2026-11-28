// Package model provides CLI commands to manage models registered with the catalog.
package model

import (
	"github.com/spf13/cobra"
)

// ModelCmd returns the top-level model command.
func ModelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Manage models",
		Long:  `Manage the models (local and remote) registered with the catalog.`,
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newInfoCmd())
	cmd.AddCommand(newPsCmd())

	return cmd
}
