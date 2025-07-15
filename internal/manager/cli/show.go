package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nekr0z/gk/internal/manager/secret"
)

func showCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show the secret",
		Long:  "Show the secret",
		Args:  cobra.ExactArgs(1),
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

	cmd.PersistentFlags().StringP("target-file", "t", "", "file to save the secret content to (otherwise will only print to stdout)")
	viper.BindPFlag("target-file", cmd.PersistentFlags().Lookup("target-file"))

	return cmd
}
