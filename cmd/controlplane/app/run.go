package app

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"time"

	"github.com/chez-shanpu/acar/pkg/utils"

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
	dataplaneAddr     string
	dataplaneTls      bool
	dataplaneCertFile string
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run controlplane grpc server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// logger
		l := grpclog.NewLoggerV2(os.Stdout, io.MultiWriter(os.Stdout, os.Stderr), os.Stderr)
		grpclog.SetLoggerV2(l)

		serverAddr := viper.GetString("controlplane.run.addr")
		tls := viper.GetBool("controlplane.run.tls")
		certFile := viper.GetString("controlplane.run.cert-path")
		keyFile := viper.GetString("controlplane.run.key-path")
		dpAddr := viper.GetString("controlplane.run.dp-addr")
		dpTls := viper.GetBool("controlplane.run.dp-tls")
		dpCertFile := viper.GetString("controlplane.run.dp-cert-path")
		err := runServer(serverAddr, tls, certFile, keyFile, dpAddr, dpTls, dpCertFile)
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

func newServer(addr string, tls bool, certFile string) *controlplaneServer {
	s := &controlplaneServer{
		dataplaneAddr:     addr,
		dataplaneTls:      tls,
		dataplaneCertFile: certFile,
	}
	return s
}

func runServer(serverAddr string, tls bool, certFile, keyFile string, dpAddr string, dpTls bool, dpCertFile string) error {
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
	api.RegisterControlPlaneServer(grpcServer, newServer(dpAddr, dpTls, dpCertFile))
	grpclog.Infof("server start: listen [%s]", serverAddr)
	if err := grpcServer.Serve(lis); err != nil {
		grpclog.Errorf("grpcServer.Serve(): %v", err)
		return err
	}
	return nil
}

func (s *controlplaneServer) RegisterSRPolicy(ctx context.Context, ai *api.AppInfo) (*types.Result, error) {
	var opts []grpc.DialOption
	if s.dataplaneTls {
		if s.dataplaneCertFile == "" {
			grpclog.Error("dp-cert file path is not set")
			return utils.NewResult(false, "dp-cert file path is not set"), errors.New("dp-cert file path is not set")
		}
		creds, err := credentials.NewClientTLSFromFile(s.dataplaneCertFile, "")
		if err != nil {
			grpclog.Errorf("failed to create TLS credentials %v", err)
			return utils.NewResult(false, "Failed to create TLS credentials"), err
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(s.dataplaneAddr, opts...)
	if err != nil {
		grpclog.Errorf("cannot connect to dataplane server: %v", err)
		return utils.NewResult(false, err.Error()), err
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			grpclog.Errorf("failed to close connection with dataplane server: %v", err)
		}
	}()

	c := api.NewDataPlaneClient(conn)
	dctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := c.ApplySRPolicy(dctx, &api.SRInfo{
		SrcAddr: ai.SrcAddr,
		DstAddr: ai.DstAddr,
		SidList: ai.SidList,
	})
	if err != nil {
		grpclog.Errorf("failed to applying SR-Policy: %v", err)
		return utils.NewResult(false, "failed to register SR-Policy"), err
	}

	return r, nil
}
