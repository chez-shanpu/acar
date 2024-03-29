package app

import (
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/chez-shanpu/acar/pkg/appagent"

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
	flags.String(appagent.AppName, "", "application name")
	flags.String(appagent.ControlplaneAddr, "localhost", "controlplane server address")
	flags.String(appagent.ControlplaneCert, "", "path to controlplane server cert file (this option is enabled when tls flag is true)")
	flags.Bool(appagent.ControlplaneTLS, false, "controlplane server tls flag")
	flags.StringSlice(appagent.DepSID, []string{}, "the sid of the departure")
	flags.String(appagent.DstAddr, "", "destination address")
	flags.String(appagent.DstSID, "", "the sid of the destination")
	flags.Int(appagent.Interval, 1, "calculation interval (sec)")
	flags.Float64(appagent.Lazy, 0, "lazy probability (float 0 to 1)")
	flags.StringP(appagent.Metrics, "", "bytes", "what metrics uses for make a graph ('ratio' and 'bytes' is now supported and default is 'bytes')")
	flags.String(appagent.MonitoringAddr, "localhost", "monitoring server address")
	flags.String(appagent.MonitoringCert, "", "path to monitoring server cert file (this option is enabled when tls flag is true)")
	flags.Bool(appagent.MonitoringTLS, false, "monitoring server tls flag")
	flags.Float64(appagent.Require, 0, "required metrics value (if 'byte' metrics is choosed, this value means required free bandwidth[bps])")
	flags.String(appagent.SrcAddr, "", "segment routing domain ingress interface address")

	_ = viper.BindPFlags(flags)

	// required
	_ = rootCmd.MarkFlagRequired(appagent.AppName)
	_ = rootCmd.MarkFlagRequired(appagent.ControlplaneAddr)
	_ = rootCmd.MarkFlagRequired(appagent.DepSID)
	_ = rootCmd.MarkFlagRequired(appagent.DstAddr)
	_ = rootCmd.MarkFlagRequired(appagent.DstSID)
	_ = rootCmd.MarkFlagRequired(appagent.Interval)
	_ = rootCmd.MarkFlagRequired(appagent.MonitoringAddr)
	_ = rootCmd.MarkFlagRequired(appagent.Require)
	_ = rootCmd.MarkFlagRequired(appagent.SrcAddr)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "app-agent",
	Short: "acar application agent",
	Version: fmt.Sprintf("acar application agent Version: %s (Revision: %s / GoVersion: %s / Compiler: %s)\n",
		Version, Revision, GoVersion, Compiler),
	Run: func(cmd *cobra.Command, args []string) {
		appagent.Config.Populate()

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		defer func() {
			signal.Stop(sigs)
		}()

		errCh := make(chan error, 1)
		ticker := time.NewTicker(time.Duration(appagent.Config.Interval) * time.Second)
		go func() {
			for {
				select {
				case <-ticker.C:
					if err := run(); err != nil {
						errCh <- err
						return
					}
				}
			}
		}()

		select {
		case sig := <-sigs:
			fmt.Printf("Finished with the signal: %v", sig)
		case err := <-errCh:
			fmt.Printf("[ERROR]: %v", err)
		}
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

func run() error {
	if isLazy() {
		log.Println("Lazy")
		return nil
	}

	nodes, err := appagent.GetSRNodesInfo()
	if err != nil {
		return err
	}

	graph, err := appagent.MakeGraph(nodes)
	if err != nil {
		return err
	}

	list, err := appagent.MakeSIDList(graph)
	if err != nil {
		return err
	}
	if list == nil {
		log.Println("No best path")
		return nil
	}

	return appagent.SendSRInfoToControlPlane(list)
}

func isLazy() bool {
	if appagent.Config.Lazy > randFloat() {
		return true
	} else {
		return false
	}
}
func randFloat() float64 {
	a, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(err)
	}
again:
	f := float64(a.Int64()) / (1 << 63)
	if f == 1 {
		goto again
	}
	return f
}
