package srnode

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/chez-shanpu/acar/pkg/grpc"

	"github.com/chez-shanpu/acar/api"
	"github.com/gosnmp/gosnmp"
	"golang.org/x/sync/errgroup"
)

const BytesToBits = 8.0
const MegaBitsToBits = 1000000
const ifHighSpeedOID = "1.3.6.1.2.1.31.1.1.1.15"
const ifHCOutOctetsOID = "1.3.6.1.2.1.31.1.1.1.10"
const ifIndexOID = "1.3.6.1.2.1.2.2.1.1"
const ifDescrOID = "1.3.6.1.2.1.2.2.1.2"

type NetworkInterface struct {
	Sid           string
	NextSids      []string
	InterfaceName string
	LinkCap       int64
}

func GatherMetricsBySNMP() ([]*api.Node, error) {
	var eg errgroup.Group
	var nodes []*api.Node

	mutex := &sync.Mutex{}
	for _, ni := range Config.NetworkInterfaces {
		ni := ni
		eg.Go(func() error {
			return func(ni *NetworkInterface) error {
				sc := newSNMPClient(Config.Addr, Config.Port, Config.SNMPUser, Config.SNMPAuthPass, Config.SNMPPrivPass)
				ifIndex, err := getInterfaceIndexByName(sc, ni.InterfaceName)
				if err != nil {
					return err
				}

				usageRatio, usageBytes, err := getInterfaceUsageBySNMP(sc, ifIndex, Config.Interval, ni.LinkCap)
				if err != nil {
					return err
				}

				node := &api.Node{
					SID:            ni.Sid,
					NextSids:       ni.NextSids,
					LinkCap:        ni.LinkCap,
					LinkUsageRatio: usageRatio,
					LinkUsageBytes: usageBytes,
				}
				mutex.Lock()
				nodes = append(nodes, node)
				mutex.Unlock()
				return nil
			}(ni)
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return nodes, nil
}

func SendToMonitoringServer(ns []*api.Node) error {
	conn, err := grpc.MakeConn(Config.MonitoringAddr, Config.TLS, Config.TLSCert)
	if err != nil {
		return err
	}
	defer conn.Close()

	c := api.NewMonitoringClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = c.RegisterNodes(ctx, &api.RegisterNodesRequest{Nodes: ns})
	return err
}

func newSNMPClient(addr string, port int, user, authPass, privPass string) *gosnmp.GoSNMP {
	return &gosnmp.GoSNMP{
		Target:        addr,
		Port:          uint16(port),
		Version:       gosnmp.Version3,
		SecurityModel: gosnmp.UserSecurityModel,
		MsgFlags:      gosnmp.AuthPriv,
		SecurityParameters: &gosnmp.UsmSecurityParameters{
			UserName:                 user,
			AuthenticationProtocol:   gosnmp.MD5,
			AuthenticationPassphrase: authPass,
			PrivacyProtocol:          gosnmp.DES,
			PrivacyPassphrase:        privPass,
		},
		Timeout: 10 * time.Second,
	}
}

func getInterfaceIndexByName(snmp *gosnmp.GoSNMP, ifName string) (int, error) {
	err := snmp.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer snmp.Conn.Close()

	res, err := snmp.WalkAll(ifIndexOID)
	if err != nil {
		return 0, fmt.Errorf("failed get interface number from snmp agent: %v", err)
	}

	var indexSlice []int
	for _, v := range res {
		if v.Type != gosnmp.Integer {
			return 0, fmt.Errorf("variable type is wrong correct %v, got %v", gosnmp.Integer, v.Type)
		}
		indexSlice = append(indexSlice, int(gosnmp.ToBigInt(v.Value).Int64()))
	}

	for _, i := range indexSlice {
		oids := []string{fmt.Sprintf("%s.%d", ifDescrOID, i)}
		res, err := snmp.Get(oids)
		if err != nil {
			return 0, fmt.Errorf("failed to get interface name from snmp agent: %v", err)
		}

		for _, variable := range res.Variables {
			if variable.Value == nil {
				continue
			}

			if variable.Type == gosnmp.OctetString || variable.Type == gosnmp.NoSuchInstance {
				if string(variable.Value.([]uint8)) == ifName {
					return i, nil
				}
			} else {
				return 0, fmt.Errorf("variable type is wrong correct %v, got %v", gosnmp.OctetString, variable.Type)
			}
		}
	}
	return 0, fmt.Errorf("no interface named %s", ifName)
}

func getInterfaceCapacity(snmp *gosnmp.GoSNMP, ifIndex int) (int64, error) {
	err := snmp.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer snmp.Conn.Close()

	oids := []string{fmt.Sprintf("%s.%d", ifHighSpeedOID, ifIndex)}
	res, err := snmp.Get(oids)
	if err != nil {
		return -1, fmt.Errorf("failed get metrics from snmp agent: %v", err)
	}

	var linkCapBits int64
	for _, variable := range res.Variables {
		if variable.Type != gosnmp.Gauge32 {
			return -1, fmt.Errorf("variable type is wrong correct %v, got %v", gosnmp.Gauge32, variable.Type)
		}
		linkCapBits = gosnmp.ToBigInt(variable.Value).Int64()
	}
	return linkCapBits, nil
}

func getInterfaceUsageBySNMP(snmp *gosnmp.GoSNMP, ifIndex, secInterval int, linkCap int64) (float64, float64, error) {
	if linkCap <= 0 {
		var err error
		linkCap, err = getInterfaceCapacity(snmp, ifIndex)
		if err != nil {
			return 0, 0, err
		}
	}
	// first
	firstUsageBytesMetric, err := getInterfaceUsageBytes(snmp, ifIndex)
	if err != nil {
		return 0, 0, err
	}
	firstGetTime := time.Now()

	// wait
	time.Sleep(time.Duration(secInterval) * time.Second)

	// second
	secondUsageBytesMetric, err := getInterfaceUsageBytes(snmp, ifIndex)
	if err != nil {
		return 0, 0, err
	}
	secondGetTime := time.Now()

	// calcurate
	dur := secondGetTime.Sub(firstGetTime).Seconds()
	ifUsageRatio, ifUsageBytes := calcInterfaceUsage(firstUsageBytesMetric, secondUsageBytesMetric, dur, linkCap*MegaBitsToBits)

	return ifUsageRatio, ifUsageBytes, nil
}

func getInterfaceUsageBytes(snmp *gosnmp.GoSNMP, ifIndex int) (int64, error) {
	err := snmp.Connect()
	if err != nil {
		return 0, err
	}
	defer snmp.Conn.Close()

	oids := []string{fmt.Sprintf("%s.%d", ifHCOutOctetsOID, ifIndex)}
	res, err := snmp.Get(oids)
	if err != nil {
		return 0, fmt.Errorf("failed get metrics from snmp agent: %v", err)
	}

	totalBytes := big.NewInt(0)
	for _, variable := range res.Variables {
		if variable.Type != gosnmp.Counter64 {
			return 0, fmt.Errorf("variable type is wrong correct %v, got %v", gosnmp.Counter64, variable.Type)
		}
		totalBytes.Add(totalBytes, gosnmp.ToBigInt(variable.Value))
	}
	return totalBytes.Int64(), nil
}

func calcInterfaceUsage(firstBytes, secondBytes int64, duration float64, linkCapBits int64) (float64, float64) {
	traficBytesDiff := secondBytes - firstBytes

	ifUsageRatio := float64(traficBytesDiff*BytesToBits*100.0) / (duration * float64(linkCapBits))
	if ifUsageRatio < 0 {
		ifUsageRatio = 0
	}

	ifUsageBytes := float64(traficBytesDiff) / duration
	if ifUsageBytes < 0 {
		ifUsageBytes = 0
	}

	return ifUsageRatio, ifUsageBytes
}
