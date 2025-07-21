package cli

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func signupCommand(loc *i18n.Localizer) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "signup",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := initClient(cmd)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.signup.signing"}))
			err = c.Signup(cmd.Context())
			if err != nil {
				return err
			}

			username := viper.GetString("server.username")

			fmt.Fprintln(cmd.OutOrStdout(), loc.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "gk.signup.success",
				TemplateData: map[string]interface{}{
					"Username": username,
				},
			}))

			return nil
		},
	}

	cmd.Short = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.signup.short"})
	cmd.Long = loc.MustLocalize(&i18n.LocalizeConfig{MessageID: "gk.signup.long"})

	return cmd
}
