package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/chez-shanpu/acar/pkg/srnode"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var (
	cfgFile string

	// Version is version number which automatically set on build. `git describe --tags`
	Version string
	// Revision is git commit hash which automatically set `git rev-parse --short HEAD` on build.
	Revision string

	GoVersion = runtime.Version()
	Compiler  = runtime.Compiler
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.acar/srnode-agent.yaml)")

	// flags
	flags := rootCmd.Flags()
	flags.String(srnode.Addr, "localhost", "SRNode address")
	flags.Int(srnode.Interval, 60, "measurement interval when measuring the interface usage rate (sec)")
	flags.String(srnode.MonitoringAddr, "localhost", "monitoring server address")
	flags.Uint16(srnode.Port, 161, "SRNode snmp port num")
	flags.String(srnode.SNMPAuthPass, "", "snmp authentication password")
	flags.String(srnode.SNMPPrivPass, "", "snmp privacy password")
	flags.String(srnode.SNMPUser, "", "snmp user name")
	flags.BoolP(srnode.TLS, "t", false, "monitoring server tls flag")
	flags.String(srnode.TLSCert, "", "path to monitoring server cert file (this option is enabled when tls flag is true)")

	// bind flags
	_ = viper.BindPFlags(flags)

	// required
	_ = rootCmd.MarkFlagRequired(srnode.Addr)
	_ = rootCmd.MarkFlagRequired(srnode.MonitoringAddr)
	_ = rootCmd.MarkFlagRequired(srnode.SNMPAuthPass)
	_ = rootCmd.MarkFlagRequired(srnode.SNMPPrivPass)
	_ = rootCmd.MarkFlagRequired(srnode.SNMPUser)

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		cfgFile = filepath.Join(home, ".acar/srnode-agent.yaml")
	}
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	viper.WatchConfig()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "srnode-agent",
	Short: "acar SRNode Agent",
	Version: fmt.Sprintf("acar srnode agent Version: %s (Revision: %s / GoVersion: %s / Compiler: %s)\n",
		Version, Revision, GoVersion, Compiler),
	RunE: func(cmd *cobra.Command, args []string) error {
		srnode.Config.Populate()
		return run()
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
	nodes, err := srnode.GatherMetricsBySNMP()
	if err != nil {
		return err
	}
	return srnode.SendToMonitoringServer(nodes)
}
