// Package cli is the command line interface for the password manager application.
package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nekr0z/gk/internal/manager/client"
	"github.com/nekr0z/gk/internal/manager/storage"
	"github.com/nekr0z/gk/internal/manager/storage/sqlite"
	"github.com/nekr0z/gk/internal/version"
)

// RootCmd returns the root command for the application.
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gk",
		Short:   "GophKeeper password manager",
		Long:    `A password manager written in Go.`,
		Version: version.String(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	cmd.PersistentFlags().StringP("db", "d", "", "database file (default is gk.sqlite in current directory)")
	viper.BindPFlag("db", cmd.PersistentFlags().Lookup("db"))
	viper.SetDefault("db", "gk.sqlite")

	cmd.PersistentFlags().StringP("passphrase", "p", "", "passphrase for encryption")
	viper.BindPFlag("passphrase", cmd.PersistentFlags().Lookup("passphrase"))

	cmd.PersistentFlags().StringP("server", "s", "", "server address")
	viper.BindPFlag("server.address", cmd.PersistentFlags().Lookup("server"))

	cmd.PersistentFlags().StringP("username", "u", "", "user name")
	viper.BindPFlag("server.username", cmd.PersistentFlags().Lookup("username"))

	cmd.PersistentFlags().StringP("password", "w", "", "password")
	viper.BindPFlag("server.password", cmd.PersistentFlags().Lookup("password"))

	cmd.PersistentFlags().BoolP("insecure", "i", false, "disable TLS verification")
	viper.BindPFlag("server.insecure", cmd.PersistentFlags().Lookup("insecure"))

	cmd.PersistentFlags().StringP("prefer", "g", "", "`remote` or `local`")
	viper.BindPFlag("prefer", cmd.PersistentFlags().Lookup("prefer"))

	cmd.PersistentFlags().StringP("config", "c", "", "config file (if not set, will look for .gk.yaml in the home directory)")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))

	viper.SetConfigName(".gk")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/")

	cmd.AddCommand(createCmd())
	cmd.AddCommand(deleteCmd())
	cmd.AddCommand(showCmd())
	cmd.AddCommand(signupCommand())
	cmd.AddCommand(syncCommand())

	return cmd
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetEnvPrefix("GK")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
	}

	viper.ReadInConfig()
}

func initStorage(cmd *cobra.Command) (*storage.Repository, error) {
	dbFilename := viper.GetString("db")
	db, err := sqlite.New(dbFilename)
	if err != nil {
		return nil, err
	}

	cmd.PostRunE = func(_ *cobra.Command, _ []string) error {
		return db.Close()
	}

	var opts []storage.Option

	if viper.GetString("server.address") != "" {
		c, err := initClient(cmd)
		if err != nil {
			return nil, err
		}

		opts = append(opts, storage.UseRemote(c))
	}

	if viper.GetString("prefer") != "" {
		switch viper.GetString("prefer") {
		case "remote":
			opts = append(opts, storage.UseResolver(storage.PreferRemote()))
		case "local":
			opts = append(opts, storage.UseResolver(storage.PreferLocal()))
		}
	}

	return storage.New(db, viper.GetString("passphrase"), opts...)
}

func initClient(cmd *cobra.Command) (*client.Client, error) {
	cfg := client.Config{
		Address:  viper.GetString("server.address"),
		Username: viper.GetString("server.username"),
		Password: viper.GetString("server.password"),
		Insecure: viper.GetBool("server.insecure"),
	}

	c, err := client.New(cmd.Context(), cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}
