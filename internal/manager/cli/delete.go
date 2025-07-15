package cli

import "github.com/spf13/cobra"

func deleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := initStorage(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			return repo.Delete(cmd.Context(), name)
		},
	}

	return cmd
}
