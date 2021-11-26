package monitoring

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/chez-shanpu/acar/pkg/logging"
	"github.com/chez-shanpu/acar/pkg/logging/logfields"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-redis/redis/v8"

	"github.com/chez-shanpu/acar/api"
)

const (
	pkg                 = "dataplane"
	nodeinfoRedisPrefix = "sids/"
	nextSidsKey         = "NextSids"
	linkCapKey          = "linkcap"
	usageRatioKey       = "ratio"
	usageBytesKey       = "bytes"
)

var log = logging.DefaultLogger.WithField(logfields.Package, pkg)

type Server struct {
	api.UnimplementedMonitoringServer
}

func NewServer() *Server {
	s := &Server{}
	return s
}

func (s *Server) GetNodes(ctx context.Context, r *api.GetNodesRequest) (*api.GetNodesResponse, error) {
	ns, err := getNodes(ctx)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &api.GetNodesResponse{Nodes: ns, Time: timestamppb.Now()}, nil
}

func (s *Server) RegisterNodes(ctx context.Context, r *api.RegisterNodesRequest) (*api.RegisterNodesResponse, error) {
	if err := registerNodes(ctx, r.GetNodes()); err != nil {
		log.WithField("Nodes", fmt.Sprintf("%#v", r.GetNodes())).Error(err)
		return nil, err
	}
	return &api.RegisterNodesResponse{Time: timestamppb.Now()}, nil
}

func getNodes(ctx context.Context) ([]*api.Node, error) {
	var ns []*api.Node

	rdb := newRedisClient(Config.RedisAddr, Config.RedisPass, Config.RedisDB)
	res := rdb.Keys(ctx, nodeinfoRedisPrefix+"*")
	keys, err := res.Result()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		sid := strings.Split(key, "/")[1]
		n := api.Node{SID: sid}
		res := rdb.HGetAll(ctx, key)
		ifInfo, err := res.Result()
		if err != nil {
			return nil, err
		}
		n.NextSids = strings.Split(ifInfo[nextSidsKey], ",")
		n.LinkCap, err = strconv.ParseInt(ifInfo[linkCapKey], 10, 64)
		if err != nil {
			return nil, err
		}
		n.LinkUsageRatio, err = strconv.ParseFloat(ifInfo[usageRatioKey], 64)
		if err != nil {
			return nil, err
		}
		n.LinkUsageBytes, err = strconv.ParseFloat(ifInfo[usageBytesKey], 64)
		if err != nil {
			return nil, err
		}
		ns = append(ns, &n)
	}
	return ns, nil
}

func registerNodes(ctx context.Context, ns []*api.Node) error {
	rdb := newRedisClient(Config.RedisAddr, Config.RedisPass, Config.RedisDB)

	for _, n := range ns {
		nextSidsStr := strings.Join(n.NextSids, ",")
		res := rdb.HSet(ctx, nodeinfoRedisPrefix+n.SID, nextSidsKey, nextSidsStr)
		if _, err := res.Result(); err != nil {
			return err
		}

		res = rdb.HSet(ctx, nodeinfoRedisPrefix+n.SID, linkCapKey, n.LinkCap)
		if _, err := res.Result(); err != nil {
			return err
		}

		res = rdb.HSet(ctx, nodeinfoRedisPrefix+n.SID, usageRatioKey, n.LinkUsageRatio)
		if _, err := res.Result(); err != nil {
			return err
		}

		res = rdb.HSet(ctx, nodeinfoRedisPrefix+n.SID, usageBytesKey, n.LinkUsageBytes)
		if _, err := res.Result(); err != nil {
			return err
		}
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
