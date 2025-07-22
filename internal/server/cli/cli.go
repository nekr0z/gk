// Package cli is the command line interface for the secrets synchronization
// server.
package cli

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/nekr0z/gk/internal/server/db"
	grpcserver "github.com/nekr0z/gk/internal/server/grpc"
	"github.com/nekr0z/gk/internal/server/secret"
	"github.com/nekr0z/gk/internal/server/user"
	"github.com/nekr0z/gk/internal/version"
	"github.com/nekr0z/gk/pkg/pb"
)

// RootCmd returns the root command for the application.
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gk-server",
		Short:   "GophKeeper server",
		Long:    `Synchronization server for GophKeeper.`,
		Version: version.String(),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := db.New(viper.GetString("dsn"))
			if err != nil {
				return err
			}
			defer db.Close()

			user, err := user.NewUserService(db, user.WithTokenSigningKey([]byte(viper.GetString("key"))))
			if err != nil {
				return err
			}

			us := grpcserver.NewUserService(user)
			secr := secret.NewService(db)
			ss := grpcserver.NewSecretServiceServer(secr)

			lis, err := net.Listen("tcp", viper.GetString("address"))
			if err != nil {
				return err
			}
			defer lis.Close()

			server := grpc.NewServer(grpc.ChainUnaryInterceptor(grpcserver.TokenInterceptor(user)))

			pb.RegisterSecretServiceServer(server, ss)
			pb.RegisterUserServiceServer(server, us)

			go func() {
				<-cmd.Context().Done()

				fmt.Fprintln(cmd.OutOrStdout(), "Shutting down...")

				server.GracefulStop()
			}()

			fmt.Fprintf(cmd.OutOrStdout(), "Running server on %s\n", lis.Addr().String())

			return server.Serve(lis)
		},
	}

	cmd.PersistentFlags().StringP("dsn", "d", "", "DSN for connection to PostgreSQL database")
	viper.BindPFlag("dsn", cmd.PersistentFlags().Lookup("dsn"))

	cmd.PersistentFlags().StringP("key", "k", "", "JWT signing key")
	viper.BindPFlag("key", cmd.PersistentFlags().Lookup("key"))

	cmd.PersistentFlags().StringP("address", "a", "", "server address")
	viper.BindPFlag("address", cmd.PersistentFlags().Lookup("address"))

	cmd.PersistentFlags().StringP("config", "c", "", "config file (if not set, will look for gk-server.yaml in the current directory)")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))

	viper.SetConfigName("gk-server")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	return cmd
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetEnvPrefix("GK_SERVER")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
	}

	viper.ReadInConfig()
}
