package appagent

import (
	"context"
	"fmt"
	"time"

	"github.com/RyanCarrier/dijkstra"
	"github.com/chez-shanpu/acar/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const significantDigits = 100000
const startVertexName = "start"

func GetSRNodesInfo(tls bool, certFilePath, mntAddr string) (*api.NodesInfo, error) {
	var opts []grpc.DialOption
	if tls {
		if certFilePath == "" {
			return nil, fmt.Errorf("dp-cert file path is not set")
		}
		creds, err := credentials.NewClientTLSFromFile(certFilePath, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(mntAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	defer conn.Close()

	c := api.NewMonitoringServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodesInfo, err := c.GetNodes(ctx, &api.GetNodesParams{})
	if err != nil {
		return nil, fmt.Errorf("RegisterSRPolicy(): %v", err)
	}
	return nodesInfo, nil
}

func MakeGraph(nodesInfo *api.NodesInfo) (*dijkstra.Graph, error) {
	graph := dijkstra.NewGraph()
	for _, node := range nodesInfo.Nodes {
		graph.AddMappedVertex(node.SID)
		for _, lc := range node.LinkCosts {
			// RyanCarrier/dijkstra's link cost can be only integer
			cost := int64(lc.Cost * significantDigits)
			err := graph.AddMappedArc(node.SID, lc.NextSid, cost)
			if err != nil {
				return nil, fmt.Errorf("graph.AddMappedArc was failed: %v", err)
			}
		}
	}
	return graph, nil
}

func MakeSIDList(graph *dijkstra.Graph, depSids []string, dstSid string) (*[]string, error) {
	dstSidIndex, err := graph.GetMapping(dstSid)
	if err != nil {
		return nil, fmt.Errorf("failed to graph.GetMapping with destination address (Is your `dst-sid` correct?): %v", err)
	}

	startID := graph.AddMappedVertex(startVertexName)
	for _, depSid := range depSids {
		err = graph.AddMappedArc(startVertexName, depSid, 0)
		if err != nil {
			return nil, fmt.Errorf("graph.AddMappedArc was failed: %v", err)
		}
	}

	best, err := graph.Shortest(startID, dstSidIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to searching shortest path: %v", err)
	}

	var sids []string
	for _, verIndex := range best.Path {
		sid, _ := graph.GetMapped(verIndex)
		if sid != startVertexName {
			sids = append(sids, sid)
		}
	}
	if sids == nil {
		return nil, fmt.Errorf("something wrong with calc route: sid list is empth")
	}
	return &sids, nil
}

func SendSRInfoToControlPlane(sidList *[]string, tls bool, certFilePath, cpAddr, appName, srcAddr, dstAddr string) error {
	var opts []grpc.DialOption
	if tls {
		if certFilePath == "" {
			return fmt.Errorf("dp-cert file path is not set")
		}
		creds, err := credentials.NewClientTLSFromFile(certFilePath, "")
		if err != nil {
			return fmt.Errorf("failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(cpAddr, opts...)
	if err != nil {
		return fmt.Errorf("did not connect: %v", err)
	}
	defer conn.Close()

	c := api.NewControlPlaneClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

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
