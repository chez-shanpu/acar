/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc/grpclog"

	"github.com/spf13/viper"

	"github.com/chez-shanpu/acar/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/gosnmp/gosnmp"
	"github.com/spf13/cobra"
)

const BytesToBits = 8.0
const ifHighSpeedOID = "1.3.6.1.2.1.31.1.1.1.15"
const ifHCInOctetsOID = "1.3.6.1.2.1.31.1.1.1.6"
const ifHCOutOctetsOID = "1.3.6.1.2.1.31.1.1.1.10"
const ifNumberOID = "1.3.6.1.2.1.2.1.0"
const ifDescrOID = "1.3.6.1.2.1.2.2.1.2"

type config struct {
	networkInterfaces []NetworkInterface
}

type NetworkInterface struct {
	Sid           string
	NextSid       string
	InterfaceName string
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run SRNode Agent",
	Run: func(cmd *cobra.Command, args []string) {
		var c config
		var nodes []*api.Node
		var mutex = &sync.Mutex{}
		var eg errgroup.Group

		// logger
		l := grpclog.NewLoggerV2(os.Stdout, io.MultiWriter(os.Stdout, os.Stderr), os.Stderr)
		grpclog.SetLoggerV2(l)

		_ = viper.Unmarshal(&c)

		srnodeAddr := viper.GetString("srnode-agent.run.srnode-addr")
		srnodePort := uint16(viper.GetUint("srnode-agent.run.srnode-port"))
		snmpUser := viper.GetString("srnode-agent.run.snmp-user")
		snmpAuthPass := viper.GetString("srnode-agent.run.snmp-auth-pass")
		snmpPrivPass := viper.GetString("srnode-agent.run.snmp-priv-pass")
		sc := newSNMPClient(srnodeAddr, srnodePort, snmpUser, snmpAuthPass, snmpPrivPass)

		interval := viper.GetInt("srnode-agent.run.interval")
		for _, ni := range c.networkInterfaces {
			ni := ni
			eg.Go(func() error {
				return func(ni NetworkInterface) error {
					ifIndex, err := getInterfaceIndexByName(sc, ni.InterfaceName)
					if err != nil {
						l.Errorf("failed to get interface index: %v", ifIndex)
					}
					usage, err := getInterfaceUsagePercentBySNMP(sc, ifIndex, interval)
					if err != nil {
						l.Errorf("failed to get interface usage: %v", err)
					}
					node := api.Node{
						SID: ni.Sid,
						LinkCosts: []*api.LinkCost{
							{
								NextSid: ni.NextSid,
								Cost:    usage,
							},
						},
					}
					mutex.Lock()
					nodes = append(nodes, &node)
					mutex.Unlock()
					return nil
				}(ni)
			})
		}

		// The link cost to different SIDs in the same host is 0
		for _, n := range nodes {
			for _, ni := range c.networkInterfaces {
				if n.SID != ni.Sid {
					lc := api.LinkCost{
						NextSid: ni.Sid,
						Cost:    0,
					}
					n.LinkCosts = append(n.LinkCosts, &lc)
				}
			}
		}

		tls := viper.GetBool("srnode-agent.run.tls")
		certFilePath := viper.GetString("srnode-agent.run.cert-path")
		mntAddr := viper.GetString("srnode-agent.run.mnt-addr")
		err := sendToMonitoringServer(tls, certFilePath, mntAddr, &api.NodesInfo{Nodes: nodes})
		if err != nil {
			l.Errorf("failed to send metrics to monitoring server: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// flags
	flags := runCmd.Flags()
	flags.StringP("config", "c", "", "path to config file")
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
	_ = viper.BindPFlag("srnode-agent.run.config", flags.Lookup("config"))
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
	_ = runCmd.MarkFlagRequired("config")
	_ = runCmd.MarkFlagRequired("mnt-addr")
	_ = runCmd.MarkFlagRequired("srnode-addr")
	_ = runCmd.MarkFlagRequired("snmp-user")
	_ = runCmd.MarkFlagRequired("snmp-auth-pass")
	_ = runCmd.MarkFlagRequired("snmp-priv-pass")
}

func newSNMPClient(addr string, port uint16, user, authPass, privPass string) *gosnmp.GoSNMP {
	return &gosnmp.GoSNMP{
		Target:        addr,
		Port:          port,
		Version:       gosnmp.Version3,
		SecurityModel: gosnmp.UserSecurityModel,
		MsgFlags:      gosnmp.AuthPriv,
		SecurityParameters: &gosnmp.UsmSecurityParameters{
			UserName:                 user,
			AuthenticationProtocol:   gosnmp.MD5,
			AuthenticationPassphrase: authPass,
			PrivacyProtocol:          gosnmp.DES,
			PrivacyPassphrase:        privPass,
		},
		Timeout: 10 * time.Second,
	}
}

func getInterfaceIndexByName(snmp *gosnmp.GoSNMP, ifName string) (int, error) {
	oids := []string{ifNumberOID}
	res, err := snmp.Get(oids)
	if err != nil {
		return -1, fmt.Errorf("failed get interface number from snmp agent: %v", err)
	}

	var maxIfIndex int
	for _, variable := range res.Variables {
		if variable.Type != gosnmp.Integer {
			return -1, fmt.Errorf("variable type is wrong: %v", variable.Type)
		}
		maxIfIndex = int(gosnmp.ToBigInt(variable.Value).Int64())
	}

	for i := 1; i <= maxIfIndex; i++ {
		oids = []string{fmt.Sprintf("%s.%d", ifDescrOID, i)}
		res, err = snmp.Get(oids)
		if err != nil {
			return -1, fmt.Errorf("failed get interface name from snmp agent: %v", err)
		}

		for _, variable := range res.Variables {
			if variable.Type != gosnmp.OctetString {
				return -1, fmt.Errorf("variable type is wrong: %v", variable.Type)
			}
			if variable.Value == ifName {
				return i, nil
			}
		}
	}
	return -1, fmt.Errorf("no interface named %s", ifName)
}

func getInterfaceUsageBytes(snmp *gosnmp.GoSNMP, ifIndex int) (int64, error) {
	oids := []string{fmt.Sprintf("%s.%d", ifHCInOctetsOID, ifIndex), fmt.Sprintf("%s.%d", ifHCOutOctetsOID, ifIndex)}
	res, err := snmp.Get(oids)
	if err != nil {
		return -1, fmt.Errorf("failed get metrics from snmp agent: %v", err)
	}

	totalBytes := big.NewInt(0)
	for _, variable := range res.Variables {
		if variable.Type != gosnmp.Counter64 {
			return -1, fmt.Errorf("variable type is wrong: %v", variable.Type)
		}
		totalBytes.Add(totalBytes, gosnmp.ToBigInt(variable.Value))
	}
	return totalBytes.Int64(), nil
}

func getInterfaceUsagePercentBySNMP(snmp *gosnmp.GoSNMP, ifIndex, secInterval int) (float64, error) {
	err := snmp.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer snmp.Conn.Close()

	// get link cap
	oids := []string{fmt.Sprintf("%s.%d", ifHighSpeedOID, ifIndex)}
	res, err := snmp.Get(oids)
	if err != nil {
		return -1, fmt.Errorf("failed get metrics from snmp agent: %v", err)
	}

	var linkCapBits int64
	for _, variable := range res.Variables {
		if variable.Type != gosnmp.Gauge32 {
			return -1, fmt.Errorf("variable type is wrong: %v", variable.Type)
		}
		linkCapBits = gosnmp.ToBigInt(variable.Value).Int64()
	}

	// first
	firstUsageBytesMetric, err := getInterfaceUsageBytes(snmp, ifIndex)
	if err != nil {
		return -1, err
	}
	firstGetTime := time.Now()

	// wait
	time.Sleep(time.Duration(secInterval) * time.Second)

	// second
	secondUsageBytesMetric, err := getInterfaceUsageBytes(snmp, ifIndex)
	if err != nil {
		return -1, err
	}
	secondGetTime := time.Now()

	// calcurate
	traficBytesDiff := secondUsageBytesMetric - firstUsageBytesMetric
	timeDiff := secondGetTime.Second() - firstGetTime.Second()
	ifUsagePercent := float64(traficBytesDiff) / (float64(timeDiff) * float64(linkCapBits)) * BytesToBits * 100.0

	return ifUsagePercent, nil
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
		return fmt.Errorf("did not connect: %v", err)
	}
	defer conn.Close()

	c := api.NewMonitoringServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.RegisterNodes(ctx, nodes)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %v", err)
	}
	if !res.Ok {
		return fmt.Errorf("failed to send metrics: %v", res.ErrStr)
	}
	return nil
}
