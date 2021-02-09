package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/RyanCarrier/dijkstra"

	"github.com/chez-shanpu/acar/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run app-agent",
	Run: func(cmd *cobra.Command, args []string) {
		mntTLSFlag := viper.GetBool("app-agent.run.mnt-tls")
		mntCertFilePath := viper.GetString("app-agent.run.mnt-cert-path")
		mntAddr := viper.GetString("app-agent.run.mnt-addr")
		nodesInfo, err := getSRNodesInfo(mntTLSFlag, mntCertFilePath, mntAddr)
		if err != nil {
			fmt.Printf("[ERROR] %v", err)
			os.Exit(1)
		}

		graph, err := makeGraph(nodesInfo)
		if err != nil {
			fmt.Printf("[ERROR] %v", err)
			os.Exit(1)
		}

		srcAddr := viper.GetString("app-agent.run.src-addr")
		dstAddr := viper.GetString("app-agent.run.dst-addr")
		topologyFilePath := viper.GetString("app-agent.run.topology-file")
		list, err := makeSIDList(graph, srcAddr, dstAddr, topologyFilePath)
		if err != nil {
			fmt.Printf("[ERROR] %v", err)
			os.Exit(1)
		}

		tls := viper.GetBool("app-agent.run.cp-tls")
		cpCertFilePath := viper.GetString("app-agent.run.cp-cert-path")
		cpAddr := viper.GetString("app-agent.run.cp-addr")
		appName := viper.GetString("app-agent.run.app-name")
		err = sendSRInfoToControlPlane(list, tls, cpCertFilePath, cpAddr, appName, srcAddr, dstAddr)
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
	flags.StringP("cp-addr", "", "localhost", "controlplane server address")
	flags.StringP("mnt-addr", "", "localhost", "monitoring server address")
	flags.BoolP("cp-tls", "", false, "controlplane server tls flag")
	flags.BoolP("mnt-tls", "", false, "monitoring server tls flag")
	flags.String("cp-cert-path", "", "path to controlplane server cert file (this option is enabled when tls flag is true)")
	flags.String("mnt-cert-path", "", "path to monitoring server cert file (this option is enabled when tls flag is true)")
	flags.String("app-name", "", "application name")
	flags.String("src-addr", "", "segment routing domain ingress interface address")
	flags.String("dst-addr", "", "destination address")
	flags.String("topology-file", "", "path to the file which is written about network topology (this option is for testing)")

	// bind flags
	_ = viper.BindPFlag("app-agent.run.cp-addr", flags.Lookup("cp-addr"))
	_ = viper.BindPFlag("app-agent.run.mnt-addr", flags.Lookup("mnt-addr"))
	_ = viper.BindPFlag("app-agent.run.cp-tls", flags.Lookup("cp-tls"))
	_ = viper.BindPFlag("app-agent.run.mnt-tls", flags.Lookup("mnt-tls"))
	_ = viper.BindPFlag("app-agent.run.cp-cert-path", flags.Lookup("cp-cert-path"))
	_ = viper.BindPFlag("app-agent.run.app-name", flags.Lookup("app-name"))
	_ = viper.BindPFlag("app-agent.run.src-addr", flags.Lookup("src-addr"))
	_ = viper.BindPFlag("app-agent.run.dst-addr", flags.Lookup("dst-addr"))
	_ = viper.BindPFlag("app-agent.run.topology-file", flags.Lookup("topology-file"))

	// required
	_ = runCmd.MarkFlagRequired("cp-addr")
	_ = runCmd.MarkFlagRequired("mnt-addr")
	_ = runCmd.MarkFlagRequired("app-name")
	_ = runCmd.MarkFlagRequired("src-addr")
	_ = runCmd.MarkFlagRequired("dst-addr")
}

func makeGraph(nodesInfo *api.NodesInfo) (*dijkstra.Graph, error) {
	graph := dijkstra.NewGraph()
	for _, node := range nodesInfo.Nodes {
		graph.AddMappedVertex(node.SID)
		for _, lc := range node.LinkCosts {
			// RyanCarrier/dijkstra's link cost can be only integer
			cost := int64(lc.Cost * 100000)
			err := graph.AddMappedArc(node.SID, lc.NextSid, cost)
			if err != nil {
				return nil, fmt.Errorf("graph.AddMappedArc was failed: %v", err)
			}
		}
	}
	return graph, nil
}

func makeSIDList(graph *dijkstra.Graph, srcAddr, dstAddr, topologyFilePath string) (*[]string, error) {
	srcAddrIndex, err := graph.GetMapping(srcAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to graph.GetMapping with source address (Is your `src-addr` correct?): %v", err)
	}
	dstAddrIndex, err := graph.GetMapping(dstAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to graph.GetMapping with destination address (Is your `dst-addr` correct?): %v", err)
	}
	best, err := graph.Shortest(srcAddrIndex, dstAddrIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to searching shortest path: %v", err)
	}

	var sids []string
	for _, verIndex := range best.Path {
		sid, _ := graph.GetMapped(verIndex)
		sids = append(sids, sid)
	}
	if sids == nil {
		return nil, fmt.Errorf("something wrong with calc route: sid list is empth")
	}
	return &sids, nil
}

func getSRNodesInfo(tls bool, certFilePath, mntAddr string) (*api.NodesInfo, error) {
	var opts []grpc.DialOption
	if tls {
		if certFilePath == "" {
			return nil, fmt.Errorf("dp-cert file path is not set")
		}
		creds, err := credentials.NewClientTLSFromFile(certFilePath, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(mntAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	defer conn.Close()

	c := api.NewMonitoringServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	nodesInfo, err := c.GetNodes(ctx, &api.GetNodesParams{})
	if err != nil {
		return nil, fmt.Errorf("RegisterSRPolicy(): %v", err)
	}
	return nodesInfo, nil
}

func sendSRInfoToControlPlane(sidList *[]string, tls bool, certFilePath, cpAddr, appName, srcAddr, dstAddr string) error {
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
	conn, err := grpc.Dial(cpAddr, opts...)
	if err != nil {
		return fmt.Errorf("did not connect: %v", err)
	}
	defer conn.Close()

	c := api.NewControlPlaneClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.RegisterSRPolicy(ctx, &api.AppInfo{
		AppName: appName,
		SrcAddr: srcAddr,
		DstAddr: dstAddr,
		SidList: *sidList,
	})
	if err != nil {
		return fmt.Errorf("RegisterSRPolicy(): %v", err)
	}
	return nil
}
