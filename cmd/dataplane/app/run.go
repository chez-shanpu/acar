package app

import (
	"context"
	"errors"
	"io"
	"net"
	"os"

	"github.com/chez-shanpu/acar/pkg/utils"

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
	devName string
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run dataplane server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// logger
		l := grpclog.NewLoggerV2(os.Stdout, io.MultiWriter(os.Stdout, os.Stderr), os.Stderr)
		grpclog.SetLoggerV2(l)

		serverAddr := viper.GetString("dataplane.run.addr")
		tls := viper.GetBool("dataplane.run.tls")
		certFile := viper.GetString("dataplane.run.cert-path")
		keyFile := viper.GetString("dataplane.run.key-path")
		devName := viper.GetString("dataplane.run.device")
		err := runServer(serverAddr, tls, certFile, keyFile, devName)
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

func newServer(devName string) *dataplaneServer {
	s := &dataplaneServer{
		devName: devName,
	}
	return s
}

func runServer(serverAddr string, tls bool, certFile, keyFile, dev string) error {
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		grpclog.Errorf("failed to listen: %v", err)
		return err
	}

	var opts []grpc.ServerOption
	if tls {
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
	api.RegisterDataPlaneServer(grpcServer, newServer(dev))
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
		sidList = append([]net.IP{net.ParseIP(sid)}, sidList...)
	}
	seg6encap.Segments = sidList

	srcIPv6Flag, err := isIPV6(si.SrcAddr)
	if err != nil {
		grpclog.Errorf("src address is wrong format: %v", err)
		return utils.NewResult(false, "src address is wrong format"), errors.New("src address is wrong format")
	}

	dstIPv6Flag, err := isIPV6(si.DstAddr)
	if err != nil {
		grpclog.Error("dst address is wrong format")
		return utils.NewResult(false, "dst address is wrong format"), errors.New("dst address is wrong format")
	}

	if srcIPv6Flag != dstIPv6Flag {
		grpclog.Error("src address and dst address are different format")
		return &types.Result{
			Ok: false,
		}, errors.New("src address and dst address are different format")
	}

	dstIP, dstIPnet, err := net.ParseCIDR(si.DstAddr)
	if err != nil {
		grpclog.Errorf("ApplySRPolicy ParseCIDR error: %v", err)
		return utils.NewResult(false, err.Error()), err
	}

	li, err := netlink.LinkByName(s.devName)
	if err != nil {
		grpclog.Errorf("failed to get Link by dev name %s: %v", s.devName, err)
		return utils.NewResult(false, "failed to get Link"), err
	}

	route := netlink.Route{
		LinkIndex: li.Attrs().Index,
		Dst: &net.IPNet{
			IP:   dstIP,
			Mask: dstIPnet.Mask,
		},
		Encap: seg6encap,
	}
	_ = netlink.RouteDel(&route)
	if err = netlink.RouteAdd(&route); err != nil {
		grpclog.Errorf("failed to add route: %v", err)
		return utils.NewResult(false, err.Error()), err
	}
	return utils.NewResult(true, ""), nil
}

func isIPV6(addr string) (bool, error) {
	for i := 0; i < len(addr); i++ {
		switch addr[i] {
		case '.':
			return false, nil
		case ':':
			return true, nil
		}
	}
	return false, errors.New("not ip addr")
}
