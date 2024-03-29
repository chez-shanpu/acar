package app

import (
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/chez-shanpu/acar/api"
	"github.com/chez-shanpu/acar/pkg/dataplane"
	"github.com/chez-shanpu/acar/pkg/grpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	flags.StringP(dataplane.ServerAddr, "a", "localhost:18080", "server address")
	flags.BoolP(dataplane.TLS, "t", false, "tls flag")
	flags.String(dataplane.TLSCert, "", "path to cert file (this option is enabled when tls flag is true)")
	flags.String(dataplane.TLSKey, "", "path to key file (this option is enabled when tls flag is true)")
	flags.StringP(dataplane.Device, "d", "", "NIC device name")

	_ = viper.BindPFlags(flags)

	// required
	_ = rootCmd.MarkFlagRequired(dataplane.Device)
	_ = rootCmd.MarkFlagRequired(dataplane.ServerAddr)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dataplane",
	Short: "acar dataplane server",
	Version: fmt.Sprintf("acar dataplane server Version: %s (Revision: %s / GoVersion: %s / Compiler: %s)\n",
		Version, Revision, GoVersion, Compiler),
	RunE: func(cmd *cobra.Command, args []string) error {
		dataplane.Config.Populate()
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
	s, err := grpc.MakeServer(dataplane.Config.TLS, dataplane.Config.TLSCert, dataplane.Config.TLSKey)
	if err != nil {
		return err
	}
	api.RegisterDataPlaneServer(s, dataplane.NewServer())

	lis, err := net.Listen("tcp", dataplane.Config.ServerAddr)
	if err != nil {
		return err
	}
	return s.Serve(lis)
}
