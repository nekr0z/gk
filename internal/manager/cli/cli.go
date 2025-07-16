// Package cli is the command line interface for the password manager application.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nekr0z/gk/internal/manager/storage"
	"github.com/nekr0z/gk/internal/manager/storage/sqlite"
	"github.com/nekr0z/gk/internal/version"
)

// Execute runs the root command with graceful shutdown on signals from OS.
func Execute() {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigChan
		cancel()
	}()

	if err := rootCmd().ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gk",
		Short:   "GophKeeper password manager",
		Long:    `A password manager written in Go.`,
		Version: version.String(),
		Run: func(cmd *cobra.Command, args []string) {
			db := viper.GetString("db")
			fmt.Println("db:", db)
		},
	}

	cmd.PersistentFlags().StringP("db", "d", "", "database file (default is gk.sqlite in current directory)")
	viper.BindPFlag("db", cmd.PersistentFlags().Lookup("db"))
	viper.SetDefault("db", "gk.sqlite")

	cmd.PersistentFlags().StringP("passphrase", "p", "", "passphrase for encryption")
	viper.BindPFlag("passphrase", cmd.PersistentFlags().Lookup("passphrase"))

	cmd.AddCommand(createCmd())
	cmd.AddCommand(deleteCmd())
	cmd.AddCommand(showCmd())

	return cmd
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetEnvPrefix("GK")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
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

	return storage.New(db, viper.GetString("passphrase"))
}
