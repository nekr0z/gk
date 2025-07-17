package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nekr0z/gk/internal/manager/secret"
)

func createCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new secret",
	}

	cmd.PersistentFlags().StringToStringP("metadata", "m", nil, "metadata for the secret (key=value), multiple can be provided")
	viper.BindPFlag("metadata", cmd.PersistentFlags().Lookup("metadata"))

	cmd.AddCommand(createTextCmd())
	cmd.AddCommand(createBinaryCmd())
	cmd.AddCommand(createPasswordCmd())
	cmd.AddCommand(createCardCmd())

	return cmd
}

func createTextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "text <name> <value>",
		Short: "Create a new text secret",
		Args:  cobra.ExactArgs(2),
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

	return cmd
}

func createBinaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "binary <name> <filename>",
		Short: "Create a new binary secret from file",
		Args:  cobra.ExactArgs(2),
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

	return cmd
}

func createPasswordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password <name> <username> <password>",
		Short: "Create a new password secret",
		Args:  cobra.ExactArgs(3),
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

	return cmd
}

func createCardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "card <name> <number> <expiry> <cvv> [<username>]",
		Short: "Create a new card secret",
		Args:  cobra.RangeArgs(4, 5),
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

	return cmd
}
