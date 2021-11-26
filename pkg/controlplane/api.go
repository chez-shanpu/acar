package controlplane

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/chez-shanpu/acar/api"
	"github.com/chez-shanpu/acar/pkg/grpc"
	"github.com/chez-shanpu/acar/pkg/logging"
	"github.com/chez-shanpu/acar/pkg/logging/logfields"
	"github.com/sirupsen/logrus"
)

const (
	pkg = "controlplane"
)

var log = logging.DefaultLogger.WithField(logfields.Package, pkg)

type Server struct {
	api.UnimplementedControlPlaneServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) RegisterSRPolicy(ctx context.Context, r *api.RegisterSRPolicyRequest) (*api.RegisterSRPolicyResponse, error) {
	if err := registerSRPolicy(r.GetAppInfo()); err != nil {
		log.WithFields(logrus.Fields{
			"AppInfo": fmt.Sprintf("%#v", r.GetAppInfo()),
		}).Error(err)
		return nil, err
	}
	return &api.RegisterSRPolicyResponse{Time: timestamppb.Now()}, nil
}

func registerSRPolicy(ai *api.AppInfo) error {
	conn, err := grpc.MakeConn(Config.DataplaneAddr, Config.DataplaneTLS, Config.DataplaneCert)
	if err != nil {
		return err
	}
	defer conn.Close()

	c := api.NewDataPlaneClient(conn)
	dctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sri := &api.SRInfo{
		SrcAddr: ai.SrcAddr,
		DstAddr: ai.DstAddr,
		SidList: ai.SidList,
	}
	_, err = c.ApplySRPolicy(dctx, &api.ApplySRPolicyRequest{SrInfo: sri})
	return err
}
