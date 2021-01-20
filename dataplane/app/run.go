package app

import (
	"context"
	"io"
	"net"
	"os"

	"github.com/chez-shanpu/acar/api"
	"github.com/chez-shanpu/acar/api/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

type dataplaneServer struct {
	api.UnimplementedDataPlaneServer
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run dataplane grpc server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// logger
		l := grpclog.NewLoggerV2(os.Stdout, io.MultiWriter(os.Stdout, os.Stderr), os.Stderr)
		grpclog.SetLoggerV2(l)
		err := runServer()
		return err
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// flags
	flags := runCmd.Flags()
	flags.StringP("addr", "a", "localhost:18080", "server address")
	flags.BoolP("tls", "t", false, "tls flag")
	flags.String("cert-path", "", "path to cert file (this option is enabled when tls flag is true)")
	flags.String("key-path", "", "path to key file (this option is enabled when tls flag is true)")
	flags.StringP("device", "d", "", "NIC device name")

	// bind flags
	_ = viper.BindPFlag("dataplane.run.addr", flags.Lookup("addr"))
	_ = viper.BindPFlag("dataplane.run.tls", flags.Lookup("tls"))
	_ = viper.BindPFlag("dataplane.run.cert-path", flags.Lookup("cert-path"))
	_ = viper.BindPFlag("dataplane.run.key-path", flags.Lookup("key-path"))
	_ = viper.BindPFlag("dataplane.run.device", flags.Lookup("device"))

	// required
	_ = runCmd.MarkFlagRequired("addr")
	_ = runCmd.MarkFlagRequired("device")
}

func newServer() *dataplaneServer {
	s := &dataplaneServer{}
	return s
}

func runServer() error {
	serverAddr := viper.GetString("dataplane.run.addr")
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		grpclog.Errorf("failed to listen: %v", err)
		return err
	}
	var opts []grpc.ServerOption

	tls := viper.GetBool("dataplane.run.tls")
	if tls {
		certFile := viper.GetString("dataplane.run.cert-path")
		keyFile := viper.GetString("dataplane.run.key-path")
		if certFile == "" || keyFile == "" {
			grpclog.Error("cert file path or key file path is not set")
			return err
		}
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			grpclog.Errorf("Failed to generate credentials %v", err)
			return err
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	api.RegisterDataPlaneServer(grpcServer, newServer())
	grpclog.Infof("server start: listen [%s]", serverAddr)
	if err := grpcServer.Serve(lis); err != nil {
		grpclog.Errorf("grpcServer.Serve(): %v", err)
		return err
	}
	return nil
}

func (s *dataplaneServer) ApplySRPolicy(ctx context.Context, si *api.SRInfo) (*types.Result, error) {
	seg6encap := &netlink.SEG6Encap{Mode: nl.SEG6_IPTUN_MODE_ENCAP}

	var sidList []net.IP
	for _, sid := range si.SidList {
		sidList = append(sidList, net.ParseIP(sid))
	}
	seg6encap.Segments = sidList

	ip, ipnet, err := net.ParseCIDR(si.DstAddr)
	if err != nil {
		grpclog.Errorf("ApplySRPolicy ParseCIDR error: %v", err)
		return &types.Result{
			Ok:     false,
			ErrStr: err.Error(),
		}, err
	}

	devName := viper.GetString("dataplane.run.device")
	li, err := netlink.LinkByName(devName)
	if err != nil {
		grpclog.Errorf("ApplySRPolicy LinkByName error: %v", err)
		return nil, err
	}
	route := netlink.Route{
		LinkIndex: li.Attrs().Index,
		Dst: &net.IPNet{
			IP:   ip,
			Mask: ipnet.Mask,
		},
		Encap: seg6encap,
	}

	if err := netlink.RouteAdd(&route); err != nil {
		grpclog.Errorf("ApplySRPolicy RouteAdd error: %v", err)
		return &types.Result{
			Ok:     false,
			ErrStr: err.Error(),
		}, err
	}
	return &types.Result{
		Ok:     true,
		ErrStr: "",
	}, nil
}
