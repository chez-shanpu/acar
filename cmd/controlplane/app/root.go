package app

import (
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/chez-shanpu/acar/api"
	"github.com/chez-shanpu/acar/pkg/controlplane"
	"github.com/chez-shanpu/acar/pkg/grpc"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var (
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
	flags.String(controlplane.DataplaneAddr, "localhost:18080", "dataplane server addr")
	flags.String(controlplane.DataplaneCert, "", "path to dataplane server cert file (this option is enabled when dp-tls flag is true)")
	flags.Bool(controlplane.DataplaneTLS, false, "dataplane client tls flag")
	flags.StringP(controlplane.ServerAddr, "a", "localhost", "server address")
	flags.BoolP(controlplane.TLS, "t", false, "tls flag")
	flags.String(controlplane.TLSCert, "", "path to cert file (this option is enabled when tls flag is true)")
	flags.String(controlplane.TLSKey, "", "path to key file (this option is enabled when tls flag is true)")

	_ = viper.BindPFlags(flags)

	// required
	_ = rootCmd.MarkFlagRequired(controlplane.DataplaneAddr)
	_ = rootCmd.MarkFlagRequired(controlplane.ServerAddr)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "controlplane",
	Short: "acar controlplane",
	Version: fmt.Sprintf("acar controlplane server Version: %s (Revision: %s / GoVersion: %s / Compiler: %s)\n",
		Version, Revision, GoVersion, Compiler),
	RunE: func(cmd *cobra.Command, args []string) error {
		controlplane.Config.Populate()
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
	s, err := grpc.MakeServer(controlplane.Config.TLS, controlplane.Config.TLSCert, controlplane.Config.TLSKey)
	if err != nil {
		return err
	}
	api.RegisterControlPlaneServer(s, controlplane.NewServer())

	lis, err := net.Listen("tcp", controlplane.Config.ServerAddr)
	if err != nil {
		return err
	}
	return s.Serve(lis)
}
