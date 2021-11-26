package dataplane

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/chez-shanpu/acar/pkg/logging"
	"github.com/chez-shanpu/acar/pkg/logging/logfields"

	"github.com/chez-shanpu/acar/api"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
)

const (
	pkg = "dataplane"
)

var log = logging.DefaultLogger.WithField(logfields.Package, pkg)

type Server struct {
	api.UnimplementedDataPlaneServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) ApplySRPolicy(ctx context.Context, r *api.ApplySRPolicyRequest) (*api.ApplySRPolicyResponse, error) {
	if err := applySRPolicy(r.GetSrInfo()); err != nil {
		log.WithFields(logrus.Fields{
			"SRInfo": fmt.Sprintf("%#v", r.GetSrInfo()),
		}).Error(err)
		return nil, err
	}
	return &api.ApplySRPolicyResponse{Time: timestamppb.Now()}, nil
}

func applySRPolicy(si *api.SRInfo) error {
	if err := validateSRInfo(si); err != nil {
		return err
	}

	li, err := netlink.LinkByName(Config.Device)
	if err != nil {
		return err
	}

	dstIP, dstIPnet, err := net.ParseCIDR(si.DstAddr)
	if err != nil {
		return err
	}

	route := netlink.Route{
		LinkIndex: li.Attrs().Index,
		Dst: &net.IPNet{
			IP:   dstIP,
			Mask: dstIPnet.Mask,
		},
		Encap: constructEncapRule(si),
	}
	_ = netlink.RouteDel(&route)
	return netlink.RouteAdd(&route)
}

func constructEncapRule(si *api.SRInfo) *netlink.SEG6Encap {
	var sidList []net.IP

	for _, sid := range si.SidList {
		sidList = append([]net.IP{net.ParseIP(sid)}, sidList...)
	}
	encap := &netlink.SEG6Encap{Mode: nl.SEG6_IPTUN_MODE_ENCAP}
	encap.Segments = sidList
	return encap
}

func validateSRInfo(si *api.SRInfo) error {
	srcIPv6Flag, err := isIPV6(si.SrcAddr)
	if err != nil {
		return errors.New("src address is wrong format")
	}

	dstIPv6Flag, err := isIPV6(si.DstAddr)
	if err != nil {
		return errors.New("dst address is wrong format")
	}

	if srcIPv6Flag != dstIPv6Flag {
		return errors.New("src address and dst address are different format")
	}
	return nil
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
