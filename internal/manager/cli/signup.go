package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func signupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signup",
		Short: "Sign up for a new account",
		Long:  "Sign up for a new account on the configured server using the configured credentials.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := initClient(cmd)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Signing up...")
			err = c.Signup(cmd.Context())
			if err != nil {
				return err
			}

			username := viper.GetString("server.username")

			fmt.Fprintf(cmd.OutOrStdout(), "Signup with username %s successful!\n", username)

			return nil
		},
	}

	return cmd
}
