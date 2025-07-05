// Package cli is the command line interface for the password manager application.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nekr0z/gk/internal/version"
)

var rootCmd = &cobra.Command{
	Use:     "gk",
	Short:   "GophKeeper password manager",
	Long:    `A password manager written in Go.`,
	Version: version.String(),
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
