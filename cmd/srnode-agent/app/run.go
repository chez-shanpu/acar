/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chez-shanpu/acar/pkg/srnode"

	"github.com/spf13/viper"

	"github.com/chez-shanpu/acar/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/spf13/cobra"
)

type config struct {
	NetworkInterfaces []*srnode.NetworkInterface
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run SRNode Agent",
	Run: func(cmd *cobra.Command, args []string) {
		// config
		var c config
		_ = viper.Unmarshal(&c)

		srnodeAddr := viper.GetString("srnode-agent.run.srnode-addr")
		srnodePort := uint16(viper.GetUint("srnode-agent.run.srnode-port"))
		snmpUser := viper.GetString("srnode-agent.run.snmp-user")
		snmpAuthPass := viper.GetString("srnode-agent.run.snmp-auth-pass")
		snmpPrivPass := viper.GetString("srnode-agent.run.snmp-priv-pass")
		interval := viper.GetInt("srnode-agent.run.interval")
		nodes, err := srnode.GatherMetricsBySNMP(c.NetworkInterfaces, interval, srnodeAddr, srnodePort, snmpUser, snmpAuthPass, snmpPrivPass)
		if err != nil {
			fmt.Printf("failed to gather metrics by snmp: %v", err)
			os.Exit(1)
		}

		tls := viper.GetBool("srnode-agent.run.tls")
		certFilePath := viper.GetString("srnode-agent.run.cert-path")
		mntAddr := viper.GetString("srnode-agent.run.mnt-addr")
		err = sendToMonitoringServer(tls, certFilePath, mntAddr, &api.NodesInfo{Nodes: nodes})
		if err != nil {
			fmt.Printf("failed to send metrics to monitoring server: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// flags
	flags := runCmd.Flags()
	flags.String("mnt-addr", "localhost", "monitoring server address")
	flags.String("srnode-addr", "localhost", "SRNode address")
	flags.Uint16("srnode-port", 161, "SRNode snmp port num")
	flags.String("snmp-user", "", "snmp user name")
	flags.String("snmp-auth-pass", "", "snmp authentication password")
	flags.String("snmp-priv-pass", "", "snmp privacy password")
	flags.BoolP("tls", "t", false, "monitoring server tls flag")
	flags.String("cert-path", "", "path to monitoring server cert file (this option is enabled when tls flag is true)")
	flags.Int("interval", 60, "measurement interval when measuring the interface usage rate (sec)")

	// bind flags
	_ = viper.BindPFlag("srnode-agent.run.mnt-addr", flags.Lookup("mnt-addr"))
	_ = viper.BindPFlag("srnode-agent.run.srnode-addr", flags.Lookup("srnode-addr"))
	_ = viper.BindPFlag("srnode-agent.run.srnode-port", flags.Lookup("srnode-port"))
	_ = viper.BindPFlag("srnode-agent.run.snmp-user", flags.Lookup("snmp-user"))
	_ = viper.BindPFlag("srnode-agent.run.snmp-auth-pass", flags.Lookup("snmp-auth-pass"))
	_ = viper.BindPFlag("srnode-agent.run.snmp-priv-pass", flags.Lookup("snmp-priv-pass"))
	_ = viper.BindPFlag("srnode-agent.run.tls", flags.Lookup("tls"))
	_ = viper.BindPFlag("srnode-agent.run.cert-path", flags.Lookup("cert-path"))
	_ = viper.BindPFlag("srnode-agent.run.interval", flags.Lookup("interval"))

	// required
	_ = runCmd.MarkFlagRequired("mnt-addr")
	_ = runCmd.MarkFlagRequired("srnode-addr")
	_ = runCmd.MarkFlagRequired("snmp-user")
	_ = runCmd.MarkFlagRequired("snmp-auth-pass")
	_ = runCmd.MarkFlagRequired("snmp-priv-pass")
}

func sendToMonitoringServer(tls bool, certFilePath, mntAddr string, nodes *api.NodesInfo) error {
	var opts []grpc.DialOption
	if tls {
		if certFilePath == "" {
			return fmt.Errorf("dp-cert file path is not set")
		}
		creds, err := credentials.NewClientTLSFromFile(certFilePath, "")
		if err != nil {
			return fmt.Errorf("failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(mntAddr, opts...)
	if err != nil {
		return fmt.Errorf("cannnot connect to monitoring-server: %v", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Printf("failed to close connection with monitoring-server: %v", err)
		}
	}()

	c := api.NewMonitoringServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = c.RegisterNodes(ctx, nodes)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %v", err)
	}
	return nil
}
