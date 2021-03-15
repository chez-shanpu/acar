package app

import (
	"context"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"github.com/go-redis/redis/v8"

	"github.com/chez-shanpu/acar/api/types"

	"github.com/chez-shanpu/acar/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/grpclog"

	"github.com/spf13/cobra"
)

const nodeinfoRedisPrefix = "sids/"
const nextSidsKey = "NextSids"
const linkCapKey = "linkcap"
const usageRatioKey = "ratio"
const usageBytesKey = "bytes"

type monitoringServer struct {
	api.UnimplementedMonitoringServerServer
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run monitoring server",
	Run: func(cmd *cobra.Command, args []string) {
		// logger
		l := grpclog.NewLoggerV2(os.Stdout, io.MultiWriter(os.Stdout, os.Stderr), os.Stderr)
		grpclog.SetLoggerV2(l)

		serverAddr := viper.GetString("monitoring.run.addr")
		tls := viper.GetBool("monitoring.run.tls")
		certFile := viper.GetString("monitoring.run.cert-path")
		keyFile := viper.GetString("monitoring.run.key-path")
		err := runServer(serverAddr, tls, certFile, keyFile)
		if err != nil {
			grpclog.Error(err)
			os.Exit(-1)
		}
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
	flags.String("redis-addr", "localhost:6379", "redis server address")
	flags.String("redis-pass", "password", "redis password")
	flags.Int("redis-db", 1, "redis db number")

	// bind flags
	_ = viper.BindPFlag("monitoring.run.addr", flags.Lookup("addr"))
	_ = viper.BindPFlag("monitoring.run.tls", flags.Lookup("tls"))
	_ = viper.BindPFlag("monitoring.run.cert-path", flags.Lookup("cert-path"))
	_ = viper.BindPFlag("monitoring.run.key-path", flags.Lookup("key-path"))
	_ = viper.BindPFlag("monitoring.run.redis-addr", flags.Lookup("redis-addr"))
	_ = viper.BindPFlag("monitoring.run.redis-pass", flags.Lookup("redis-pass"))
	_ = viper.BindPFlag("monitoring.run.redis-db", flags.Lookup("redis-db"))

	// required
	_ = runCmd.MarkFlagRequired("addr")
}

func newServer() *monitoringServer {
	s := &monitoringServer{}
	return s
}

func runServer(serverAddr string, tls bool, certFile, keyFile string) error {
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
	api.RegisterMonitoringServerServer(grpcServer, newServer())
	grpclog.Infof("server start: listen [%s]", serverAddr)
	if err := grpcServer.Serve(lis); err != nil {
		grpclog.Errorf("failed to start monitoring sever: %v", err)
		return err
	}
	return nil
}

func newRedisClient(addr, pass string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db, // use default DB
	})
}

func (s *monitoringServer) GetNodes(ctx context.Context, parm *api.GetNodesParams) (*api.NodesInfo, error) {
	redisAddr := viper.GetString("monitoring.run.redis-addr")
	redisPass := viper.GetString("monitoring.run.redis-pass")
	redisDB := viper.GetInt("monitoring.run.redis-db")
	rdb := newRedisClient(redisAddr, redisPass, redisDB)

	res := rdb.Keys(ctx, nodeinfoRedisPrefix+"*")
	keys, err := res.Result()
	if err != nil {
		grpclog.Errorf("failed to get keys: %v", err)
		return nil, err
	}

	ni := api.NodesInfo{}
	for _, key := range keys {
		sid := strings.Split(key, "/")[1]
		n := api.Node{SID: sid}
		res := rdb.HGetAll(ctx, key)
		ifInfo, err := res.Result()
		if err != nil {
			grpclog.Errorf("failed to get values with key %s: %v", key, err)
			return nil, err
		}
		n.NextSids = strings.Split(ifInfo[nextSidsKey], ",")
		n.LinkCap, err = strconv.ParseInt(ifInfo[linkCapKey], 10, 64)
		if err != nil {
			grpclog.Errorf("failed to parse linkcap: %v", err)
			return nil, err
		}
		n.LinkUsageRatio, err = strconv.ParseFloat(ifInfo[usageRatioKey], 64)
		if err != nil {
			grpclog.Errorf("failed to parse usage-ratio: %v", err)
			return nil, err
		}
		n.LinkUsageBytes, err = strconv.ParseFloat(ifInfo[usageBytesKey], 64)
		if err != nil {
			grpclog.Errorf("failed to parse usage-bytes: %v", err)
			return nil, err
		}
		ni.Nodes = append(ni.Nodes, &n)
	}

	return &ni, nil
}

func (s *monitoringServer) RegisterNodes(ctx context.Context, ni *api.NodesInfo) (*types.Result, error) {
	redisAddr := viper.GetString("monitoring.run.redis-addr")
	redisPass := viper.GetString("monitoring.run.redis-pass")
	redisDB := viper.GetInt("monitoring.run.redis-db")
	rdb := newRedisClient(redisAddr, redisPass, redisDB)

	for _, node := range ni.Nodes {
		nextSidsStr := strings.Join(node.NextSids, ",")
		res := rdb.HSet(ctx, nodeinfoRedisPrefix+node.SID, nextSidsKey, nextSidsStr)
		_, err := res.Result()
		if err != nil {
			grpclog.Errorf("failed to set node-info: %v", err)
		}

		res = rdb.HSet(ctx, nodeinfoRedisPrefix+node.SID, linkCapKey, node.LinkCap)
		_, err = res.Result()
		if err != nil {
			grpclog.Errorf("failed to set node-info: %v", err)
		}

		res = rdb.HSet(ctx, nodeinfoRedisPrefix+node.SID, usageRatioKey, node.LinkUsageRatio)
		_, err = res.Result()
		if err != nil {
			grpclog.Errorf("failed to set node-info: %v", err)
		}

		res = rdb.HSet(ctx, nodeinfoRedisPrefix+node.SID, usageBytesKey, node.LinkUsageBytes)
		_, err = res.Result()
		if err != nil {
			grpclog.Errorf("failed to set node-info: %v", err)
		}
	}
	return &types.Result{Ok: true}, nil
}
