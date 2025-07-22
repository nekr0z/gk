package cli

import (
	"fmt"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nekr0z/gk/internal/manager/secret"
)

func showCmd(loc *i18n.Localizer) *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := initStorage(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			sec, err := repo.Read(cmd.Context(), name)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), sec)

			filename := viper.GetString("target-file")
			if filename == "" {
				return nil
			}

			var bb []byte

			switch v := sec.Value().(type) {
			case *secret.Binary:
				bb = v.Bytes()
			default:
				bb = []byte(v.String())
			}

			return os.WriteFile(filename, bb, 0644)
		},
	}

	cmd.Use = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.show.use"})
	cmd.Short = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.show.short"})

	cmd.PersistentFlags().StringP("target-file", "t", "", loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.show.flags.target-file"}))
	viper.BindPFlag("target-file", cmd.PersistentFlags().Lookup("target-file"))

	return cmd
}
