package app

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"time"

	"github.com/spf13/viper"

	"github.com/chez-shanpu/acar/api"
	"github.com/chez-shanpu/acar/api/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	"github.com/spf13/cobra"
)

type controlplaneServer struct {
	api.UnimplementedControlPlaneServer
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run controlplane grpc server",
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
	flags.StringP("addr", "a", "localhost", "server address")
	flags.BoolP("tls", "t", false, "tls flag")
	flags.String("cert-path", "", "path to cert file (this option is enabled when tls flag is true)")
	flags.String("key-path", "", "path to key file (this option is enabled when tls flag is true)")
	flags.String("dp-addr", "localhost:18080", "dataplane server addr")
	flags.Bool("dp-tls", false, "dataplane client tls flag")
	flags.String("dp-cert-path", "", "path to dataplane server cert file (this option is enabled when dp-tls flag is true)")

	// bind flags
	_ = viper.BindPFlag("controlplane.run.addr", flags.Lookup("addr"))
	_ = viper.BindPFlag("controlplane.run.tls", flags.Lookup("tls"))
	_ = viper.BindPFlag("controlplane.run.cert-path", flags.Lookup("cert-path"))
	_ = viper.BindPFlag("controlplane.run.key-path", flags.Lookup("key-path"))
	_ = viper.BindPFlag("controlplane.run.dp-addr", flags.Lookup("dp-addr"))
	_ = viper.BindPFlag("controlplane.run.dp-tls", flags.Lookup("dp-tls"))
	_ = viper.BindPFlag("controlplane.run.dp-cert-path", flags.Lookup("dp-cert-path"))

	// required
	_ = runCmd.MarkFlagRequired("addr")
	_ = runCmd.MarkFlagRequired("dp-addr")
}

func newServer() *controlplaneServer {
	s := &controlplaneServer{}
	return s
}

func runServer() error {
	serverAddr := viper.GetString("controlplane.run.addr")
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		grpclog.Errorf("failed to listen: %v", err)
		return err
	}
	var opts []grpc.ServerOption

	tls := viper.GetBool("controlplane.run.tls")
	if tls {
		certFile := viper.GetString("controlplane.run.cert-path")
		keyFile := viper.GetString("controlplane.run.key-path")
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
	api.RegisterControlPlaneServer(grpcServer, newServer())
	grpclog.Infof("server start: listen [%s]", serverAddr)
	if err := grpcServer.Serve(lis); err != nil {
		grpclog.Errorf("grpcServer.Serve(): %v", err)
		return err
	}
	return nil
}

func (s *controlplaneServer) RegisterSRPolicy(ctx context.Context, ai *api.AppInfo) (*types.Result, error) {
	var opts []grpc.DialOption
	dpTls := viper.GetBool("controlplane.run.dp-tls")
	if dpTls {
		caFile := viper.GetString("controlplane.run.dp-cert-path")
		if caFile == "" {
			grpclog.Error("dp-cert file path is not set")
			return nil, errors.New("dp-cert file path is not set")
		}
		creds, err := credentials.NewClientTLSFromFile(caFile, "")
		if err != nil {
			grpclog.Errorf("Failed to create TLS credentials %v", err)
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	dpAddr := viper.GetString("controlplane.run.dp-addr")
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(dpAddr, opts...)
	if err != nil {
		grpclog.Errorf("did not connect: %v", err)
		return nil, err
	}
	defer conn.Close()

	c := api.NewDataPlaneClient(conn)

	dctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.ApplySRPolicy(dctx, &api.SRInfo{
		SrcAddr: ai.SrcAddr,
		DstAddr: ai.DstAddr,
		SidList: ai.SidList,
	})
	if err != nil {
		grpclog.Errorf("RegisterSRPolicy(): %v", err)
		return nil, err
	}

	return r, nil
}
