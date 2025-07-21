package cli

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

func syncCommand(loc *i18n.Localizer) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "sync",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := initStorage(cmd)
			if err != nil {
				return err
			}
			return repo.SyncAll(cmd.Context())
		},
	}

	cmd.Short = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.sync.short"})

	return cmd
}
