package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chez-shanpu/acar/api"

	"github.com/RyanCarrier/dijkstra"
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
		list, err := makeSIDList()
		if err != nil {
			fmt.Printf("[ERROR] %v", err)
			os.Exit(1)
		}

		err = sendSRInfoToControlPlane(list)
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
	flags.StringP("cp-addr", "a", "localhost", "controlplane server address")
	flags.BoolP("tls", "t", false, "tls flag")
	flags.String("cert-path", "", "path to controlplane server cert file (this option is enabled when tls flag is true)")
	flags.String("app-name", "", "application name")
	flags.String("src-addr", "", "segment routing domain ingress interface address")
	flags.String("dst-addr", "", "destination address")
	flags.String("topology-file", "", "path to the file which is written about network topology (this option is for testing)")

	// bind flags
	_ = viper.BindPFlag("app-agent.run.cp-addr", flags.Lookup("cp-addr"))
	_ = viper.BindPFlag("app-agent.run.tls", flags.Lookup("tls"))
	_ = viper.BindPFlag("app-agent.run.cert-path", flags.Lookup("cert-path"))
	_ = viper.BindPFlag("app-agent.run.app-name", flags.Lookup("app-name"))
	_ = viper.BindPFlag("app-agent.run.src-addr", flags.Lookup("src-addr"))
	_ = viper.BindPFlag("app-agent.run.dst-addr", flags.Lookup("dst-addr"))
	_ = viper.BindPFlag("app-agent.run.topology-file", flags.Lookup("topology-file"))

	// required
	_ = runCmd.MarkFlagRequired("cp-addr")
	_ = runCmd.MarkFlagRequired("app-name")
	_ = runCmd.MarkFlagRequired("src-addr")
	_ = runCmd.MarkFlagRequired("dst-addr")
}

func makeSIDList() (*[]string, error) {
	srcAddr := viper.GetString("app-agent.run.src-addr")
	dstAddr := viper.GetString("app-agent.run.dst-addr")

	// TODO implement making graph function (Currently, it's temporarily reading from a file.)
	topologyFile := viper.GetString("app-agent.run.topology-file")
	graph, err := dijkstra.Import(topologyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to import graph from file : %v", err)
	}

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

func sendSRInfoToControlPlane(sidList *[]string) error {
	var opts []grpc.DialOption
	tls := viper.GetBool("app-agent.run.tls")
	if tls {
		caFile := viper.GetString("app-agent.run.dp-cert-path")
		if caFile == "" {
			return fmt.Errorf("dp-cert file path is not set")
		}
		creds, err := credentials.NewClientTLSFromFile(caFile, "")
		if err != nil {
			return fmt.Errorf("failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	cpAddr := viper.GetString("app-agent.run.cp-addr")
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(cpAddr, opts...)
	if err != nil {
		return fmt.Errorf("did not connect: %v", err)
	}
	defer conn.Close()

	c := api.NewControlPlaneClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	appName := viper.GetString("app-agent.run.app-name")
	srcAddr := viper.GetString("app-agent.run.src-addr")
	dstAddr := viper.GetString("app-agent.run.dst-addr")
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
