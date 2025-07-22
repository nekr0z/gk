package cli

import (
	"fmt"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nekr0z/gk/internal/manager/secret"
)

func createCmd(loc *i18n.Localizer) *cobra.Command {
	cmd := &cobra.Command{
		Use: "create",
	}

	cmd.Short = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.create.short"})

	cmd.PersistentFlags().StringToStringP("metadata", "m", nil, loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.create.flags.metadata"}))
	viper.BindPFlag("metadata", cmd.PersistentFlags().Lookup("metadata"))

	cmd.AddCommand(createTextCmd(loc))
	cmd.AddCommand(createBinaryCmd(loc))
	cmd.AddCommand(createPasswordCmd(loc))
	cmd.AddCommand(createCardCmd(loc))

	return cmd
}

func createTextCmd(loc *i18n.Localizer) *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := initStorage(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			value := args[1]

			sec := secret.NewText(value)
			sec.SetMetadata(viper.GetStringMapString("metadata"))

			err = repo.Create(cmd.Context(), name, sec)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created secret %s\n", name)
			return nil
		},
	}

	cmd.Use = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.create.text.use"})
	cmd.Short = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.create.text.short"})

	return cmd
}

func createBinaryCmd(loc *i18n.Localizer) *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := initStorage(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			file := args[1]

			bb, err := os.ReadFile(file)
			if err != nil {
				return err
			}

			sec := secret.NewBinary(bb)
			sec.SetMetadata(viper.GetStringMapString("metadata"))

			return repo.Create(cmd.Context(), name, sec)
		},
	}

	cmd.Use = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.create.binary.use"})
	cmd.Short = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.create.binary.short"})

	return cmd
}

func createPasswordCmd(loc *i18n.Localizer) *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := initStorage(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			username := args[1]
			password := args[2]

			sec := secret.NewPassword(username, password)
			sec.SetMetadata(viper.GetStringMapString("metadata"))

			return repo.Create(cmd.Context(), name, sec)
		},
	}

	cmd.Use = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.create.password.use"})
	cmd.Short = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.create.password.short"})

	return cmd
}

func createCardCmd(loc *i18n.Localizer) *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.RangeArgs(4, 5),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := initStorage(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			number := args[1]
			expiry := args[2]
			cvv := args[3]
			username := ""
			if len(args) > 4 {
				username = args[4]
			}

			sec := secret.NewCard(number, expiry, cvv, username)
			sec.SetMetadata(viper.GetStringMapString("metadata"))

			return repo.Create(cmd.Context(), name, sec)
		},
	}

	cmd.Use = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.create.card.use"})
	cmd.Short = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.create.card.short"})

	return cmd
}
