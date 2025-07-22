package cli

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

func deleteCmd(loc *i18n.Localizer) *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := initStorage(cmd)
			if err != nil {
				return err
			}

			name := args[0]

			err = repo.Delete(cmd.Context(), name)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted secret %s\n", name)

			return nil
		},
	}

	cmd.Use = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.delete.use"})
	cmd.Short = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.delete.short"})

	return cmd
}
