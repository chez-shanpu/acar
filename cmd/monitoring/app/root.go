package app

import (
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/chez-shanpu/acar/api"
	"github.com/chez-shanpu/acar/pkg/grpc"
	"github.com/chez-shanpu/acar/pkg/monitoring"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// set these value as go build -ldflags option
	// Version is version number which automatically set on build. `git describe --tags`
	Version string
	// Revision is git commit hash which automatically set `git rev-parse --short HEAD` on build.
	Revision string

	GoVersion = runtime.Version()
	Compiler  = runtime.Compiler
)

func init() {
	// flags
	flags := rootCmd.Flags()
	flags.String(monitoring.RedisAddr, "localhost:6379", "redis server address")
	flags.String(monitoring.RedisPass, "password", "redis password")
	flags.Int(monitoring.RedisDB, 1, "redis db number")
	flags.BoolP(monitoring.TLS, "t", false, "tls flag")
	flags.String(monitoring.TLSCert, "", "path to cert file (this option is enabled when tls flag is true)")
	flags.String(monitoring.TLSKey, "", "path to key file (this option is enabled when tls flag is true)")
	flags.StringP(monitoring.ServerAddr, "a", "localhost", "server address")

	_ = viper.BindPFlags(flags)

	// required
	_ = rootCmd.MarkFlagRequired("addr")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "monitoring-server",
	Short: "acar monitoring server",
	Version: fmt.Sprintf("acar monitoring server Version: %s (Revision: %s / GoVersion: %s / Compiler: %s)\n",
		Version, Revision, GoVersion, Compiler),
	RunE: func(cmd *cobra.Command, args []string) error {
		monitoring.Config.Populate()
		return runServer()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runServer() error {
	s, err := grpc.MakeServer(monitoring.Config.TLS, monitoring.Config.TLSCert, monitoring.Config.TLSKey)
	if err != nil {
		return err
	}
	api.RegisterMonitoringServer(s, monitoring.NewServer())

	lis, err := net.Listen("tcp", monitoring.Config.ServerAddr)
	if err != nil {
		return err
	}
	return s.Serve(lis)
}
