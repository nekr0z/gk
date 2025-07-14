// Package cli is the command line interface for the password manager application.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nekr0z/gk/internal/version"
)

var rootCmd = &cobra.Command{
	Use:     "gk",
	Short:   "GophKeeper password manager",
	Long:    `A password manager written in Go.`,
	Version: version.String(),
	Run: func(cmd *cobra.Command, args []string) {
		db := viper.GetString("db")
		fmt.Println("db:", db)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("db", "d", "", "database file (default is gk.sqlite in current directory)")
	viper.BindPFlag("db", rootCmd.PersistentFlags().Lookup("db"))
	viper.SetDefault("db", "gk.sqlite")
}

func initConfig() {
	viper.SetEnvPrefix("GK")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}
