package app

import (
	"fmt"
	"os"

	"github.com/chez-shanpu/acar/pkg/appagent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run app-agent",
	Run: func(cmd *cobra.Command, args []string) {
		mntTLSFlag := viper.GetBool("app-agent.run.mnt-tls")
		mntCertFilePath := viper.GetString("app-agent.run.mnt-cert-path")
		mntAddr := viper.GetString("app-agent.run.mnt-addr")
		nodesInfo, err := appagent.GetSRNodesInfo(mntTLSFlag, mntCertFilePath, mntAddr)
		if err != nil {
			fmt.Printf("[ERROR] %v", err)
			os.Exit(1)
		}

		metricsType := viper.GetString("app-agent.run.metrics")
		require := viper.GetFloat64("app-agent.run.require")
		graph, err := appagent.MakeGraph(nodesInfo, metricsType, require)
		if err != nil {
			fmt.Printf("[ERROR] %v", err)
			os.Exit(1)
		}

		depSids := viper.GetStringSlice("app-agent.run.dep-sid")
		dstSid := viper.GetString("app-agent.run.dst-sid")
		list, err := appagent.MakeSIDList(graph, depSids, dstSid)
		if err != nil {
			fmt.Printf("[ERROR] %v", err)
			os.Exit(1)
		}

		tls := viper.GetBool("app-agent.run.cp-tls")
		cpCertFilePath := viper.GetString("app-agent.run.cp-cert-path")
		cpAddr := viper.GetString("app-agent.run.cp-addr")
		appName := viper.GetString("app-agent.run.app-name")
		srcAddr := viper.GetString("app-agent.run.src-addr")
		dstAddr := viper.GetString("app-agent.run.dst-addr")
		err = appagent.SendSRInfoToControlPlane(list, tls, cpCertFilePath, cpAddr, appName, srcAddr, dstAddr)
		if err != nil {
			fmt.Printf("[ERROR] %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// flags
	flags := runCmd.Flags()
	flags.String("cp-addr", "localhost", "controlplane server address")
	flags.String("mnt-addr", "localhost", "monitoring server address")
	flags.Bool("cp-tls", false, "controlplane server tls flag")
	flags.Bool("mnt-tls", false, "monitoring server tls flag")
	flags.String("cp-cert-path", "", "path to controlplane server cert file (this option is enabled when tls flag is true)")
	flags.String("mnt-cert-path", "", "path to monitoring server cert file (this option is enabled when tls flag is true)")
	flags.String("app-name", "", "application name")
	flags.String("src-addr", "", "segment routing domain ingress interface address")
	flags.String("dst-addr", "", "destination address")
	flags.StringSlice("dep-sid", []string{}, "the sid of the departure")
	flags.String("dst-sid", "", "the sid of the destination")
	flags.StringP("metrics", "", "bits", "what metrics uses for make a graph ('ratio' and 'bits' is now supported and default is 'bits')")
	flags.Float64("require", 0, "required metrics value (if 'bits' metrics is choosed, this value means required free bandwidth[bps])")

	// bind flags
	_ = viper.BindPFlag("app-agent.run.cp-addr", flags.Lookup("cp-addr"))
	_ = viper.BindPFlag("app-agent.run.mnt-addr", flags.Lookup("mnt-addr"))
	_ = viper.BindPFlag("app-agent.run.cp-tls", flags.Lookup("cp-tls"))
	_ = viper.BindPFlag("app-agent.run.mnt-tls", flags.Lookup("mnt-tls"))
	_ = viper.BindPFlag("app-agent.run.cp-cert-path", flags.Lookup("cp-cert-path"))
	_ = viper.BindPFlag("app-agent.run.app-name", flags.Lookup("app-name"))
	_ = viper.BindPFlag("app-agent.run.src-addr", flags.Lookup("src-addr"))
	_ = viper.BindPFlag("app-agent.run.dst-addr", flags.Lookup("dst-addr"))
	_ = viper.BindPFlag("app-agent.run.dep-sid", flags.Lookup("dep-sid"))
	_ = viper.BindPFlag("app-agent.run.dst-sid", flags.Lookup("dst-sid"))
	_ = viper.BindPFlag("app-agent.run.metrics", flags.Lookup("metrics"))
	_ = viper.BindPFlag("app-agent.run.require", flags.Lookup("require"))

	// required
	_ = runCmd.MarkFlagRequired("cp-addr")
	_ = runCmd.MarkFlagRequired("mnt-addr")
	_ = runCmd.MarkFlagRequired("app-name")
	_ = runCmd.MarkFlagRequired("src-addr")
	_ = runCmd.MarkFlagRequired("dst-addr")
	_ = runCmd.MarkFlagRequired("dep-sid")
	_ = runCmd.MarkFlagRequired("dst-sid")
	_ = runCmd.MarkFlagRequired("require")
}
