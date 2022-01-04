package appagent

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/RyanCarrier/dijkstra"
	"github.com/chez-shanpu/acar/api"
	"github.com/chez-shanpu/acar/pkg/grpc"
)

const (
	startVertexName        = "start"
	ratioMetricsTypeOption = "ratio"
	bytesMetricsTypeOption = "bytes"
	infCost                = 999999
	byteToBit              = 8.0
)

func GetSRNodesInfo() ([]*api.Node, error) {
	conn, err := grpc.MakeConn(Config.MonitoringAddr, Config.MonitoringTLS, Config.MonitoringCert)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	c := api.NewMonitoringClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := c.GetNodes(ctx, &api.GetNodesRequest{})
	if err != nil {
		return nil, err
	}
	return res.GetNodes(), nil
}

func MakeGraph(nodes []*api.Node) (*dijkstra.Graph, error) {
	graph := dijkstra.NewGraph()
	for _, node := range nodes {
		graph.AddMappedVertex(node.SID)
		cost, err := makeCost(node)
		if err != nil {
			return nil, err
		}

		for _, ns := range node.NextSids {
			if err = graph.AddMappedArc(node.SID, ns, cost); err != nil {
				return nil, err
			}
		}
	}
	return graph, nil
}

func MakeSIDList(g *dijkstra.Graph) ([]string, error) {
	best, err := calcPath(g, Config.DepSIDs, Config.DstSID)
	if err != nil {
		return nil, err
	}
	return constructSIDList(g, best), nil
}

func SendSRInfoToControlPlane(sidList []string) error {
	conn, err := grpc.MakeConn(Config.ControlplaneAddr, Config.ControlplaneTLS, Config.ControlplaneCert)
	if err != nil {
		return err
	}
	defer conn.Close()

	c := api.NewControlPlaneClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ai := &api.AppInfo{
		AppName: Config.AppName,
		SrcAddr: Config.SrcAddr,
		DstAddr: Config.DstAddr,
		SidList: sidList,
	}
	_, err = c.RegisterSRPolicy(ctx, &api.RegisterSRPolicyRequest{AppInfo: ai})
	return err
}

func makeCost(node *api.Node) (int64, error) {
	cost := int64(0)

	switch Config.MetricsType {
	case ratioMetricsTypeOption:
		if Config.RequireValue <= (100 - node.LinkUsageRatio) {
			cost = 1
		} else {
			cost = infCost
		}
	case bytesMetricsTypeOption:
		if Config.RequireValue <= float64(node.LinkCap)-node.LinkUsageBytes*byteToBit {
			cost = 1
		} else {
			cost = infCost
		}
		log.WithFields(logrus.Fields{
			"Node":    fmt.Sprintf("%#v", node),
			"require": Config.RequireValue,
			"cost":    cost,
		}).Info("node cost is calculated by bytes metrics")
	default:
		return 0, fmt.Errorf("metrics type is wrong: %s", Config.MetricsType)
	}
	return cost, nil
}

func calcPath(g *dijkstra.Graph, deps []string, dst string) (*dijkstra.BestPath, error) {
	dstSidIndex, err := g.GetMapping(dst)
	if err != nil {
		return nil, err
	}
	startID := g.AddMappedVertex(startVertexName)
	for _, dep := range deps {
		if err = g.AddMappedArc(startVertexName, dep, 0); err != nil {
			return nil, err
		}
	}
	best, err := g.Shortest(startID, dstSidIndex)
	return &best, err
}

func constructSIDList(g *dijkstra.Graph, b *dijkstra.BestPath) []string {
	var sids []string

	for _, verIndex := range b.Path {
		sid, _ := g.GetMapped(verIndex)
		if sid != startVertexName {
			sids = append(sids, sid)
		}
	}
	return sids
}
