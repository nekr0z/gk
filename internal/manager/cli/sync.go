package cli

import (
	"github.com/spf13/cobra"
)

func syncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync secrets with the server",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := initStorage(cmd)
			if err != nil {
				return err
			}
			return repo.SyncAll(cmd.Context())
		},
	}

	return cmd
}
