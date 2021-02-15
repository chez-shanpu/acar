package srnode

import (
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/chez-shanpu/acar/api"
	"github.com/gosnmp/gosnmp"
	"golang.org/x/sync/errgroup"
)

const BytesToBits = 8.0
const ifHighSpeedOID = "1.3.6.1.2.1.31.1.1.1.15"
const ifHCInOctetsOID = "1.3.6.1.2.1.31.1.1.1.6"
const ifHCOutOctetsOID = "1.3.6.1.2.1.31.1.1.1.10"
const ifNumberOID = "1.3.6.1.2.1.2.1.0"
const ifDescrOID = "1.3.6.1.2.1.2.2.1.2"

type NetworkInterface struct {
	Sid           string
	NextSid       string
	InterfaceName string
}

func GatherMetricsBySNMP(networkInterfaces []*NetworkInterface, sc *gosnmp.GoSNMP, interval int) ([]*api.Node, error) {
	var eg errgroup.Group
	var nodes []*api.Node

	mutex := &sync.Mutex{}
	for _, ni := range networkInterfaces {
		ni := ni
		eg.Go(func() error {
			return func(ni *NetworkInterface) error {
				ifIndex, err := getInterfaceIndexByName(sc, ni.InterfaceName)
				if err != nil {
					return err
				}
				usage, err := getInterfaceUsagePercentBySNMP(sc, ifIndex, interval)
				if err != nil {
					return err
				}
				node := api.Node{
					SID:       ni.Sid,
					LinkCosts: []*api.LinkCost{NewLinkCost(ni.NextSid, usage)},
				}
				mutex.Lock()
				nodes = append(nodes, &node)
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

func NewLinkCost(nextSid string, cost float64) *api.LinkCost {
	return &api.LinkCost{
		NextSid: nextSid,
		Cost:    cost,
	}
}

func getInterfaceIndexByName(snmp *gosnmp.GoSNMP, ifName string) (int, error) {
	err := snmp.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer snmp.Conn.Close()

	oids := []string{ifNumberOID}
	res, err := snmp.Get(oids)
	if err != nil {
		return 0, fmt.Errorf("failed get interface number from snmp agent: %v", err)
	}

	var maxIfIndex int
	for _, variable := range res.Variables {
		if variable.Type != gosnmp.Integer {
			return 0, fmt.Errorf("variable type is wrong: %v", variable.Type)
		}
		maxIfIndex = int(gosnmp.ToBigInt(variable.Value).Int64())
	}

	for i := 1; i <= maxIfIndex; i++ {
		oids = []string{fmt.Sprintf("%s.%d", ifDescrOID, i)}
		res, err = snmp.Get(oids)
		if err != nil {
			return 0, fmt.Errorf("failed get interface name from snmp agent: %v", err)
		}

		for _, variable := range res.Variables {
			if variable.Type != gosnmp.OctetString {
				return 0, fmt.Errorf("variable type is wrong: %v", variable.Type)
			}
			if variable.Value == ifName {
				return i, nil
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
			return -1, fmt.Errorf("variable type is wrong: %v", variable.Type)
		}
		linkCapBits = gosnmp.ToBigInt(variable.Value).Int64()
	}
	return linkCapBits, nil
}

func getInterfaceUsageBytes(snmp *gosnmp.GoSNMP, ifIndex int) (int64, error) {
	err := snmp.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer snmp.Conn.Close()

	oids := []string{fmt.Sprintf("%s.%d", ifHCInOctetsOID, ifIndex), fmt.Sprintf("%s.%d", ifHCOutOctetsOID, ifIndex)}
	res, err := snmp.Get(oids)
	if err != nil {
		return 0, fmt.Errorf("failed get metrics from snmp agent: %v", err)
	}

	totalBytes := big.NewInt(0)
	for _, variable := range res.Variables {
		if variable.Type != gosnmp.Counter64 {
			return 0, fmt.Errorf("variable type is wrong: %v", variable.Type)
		}
		totalBytes.Add(totalBytes, gosnmp.ToBigInt(variable.Value))
	}
	return totalBytes.Int64(), nil
}

func calcInterfaceUsagePercent(firstBytes, secondBytes int64, firstTime, secondTime int, linkCapBits int64) float64 {
	traficBytesDiff := secondBytes - firstBytes
	timeDiff := secondTime - firstTime
	ifUsagePercent := float64(traficBytesDiff) / (float64(timeDiff) * float64(linkCapBits)) * BytesToBits * 100.0
	return ifUsagePercent
}

func getInterfaceUsagePercentBySNMP(snmp *gosnmp.GoSNMP, ifIndex, secInterval int) (float64, error) {
	linkCapBits, err := getInterfaceCapacity(snmp, ifIndex)
	if err != nil {
		return 0, err
	}
	// first
	firstUsageBytesMetric, err := getInterfaceUsageBytes(snmp, ifIndex)
	if err != nil {
		return 0, err
	}
	firstGetTime := time.Now()

	// wait
	time.Sleep(time.Duration(secInterval) * time.Second)

	// second
	secondUsageBytesMetric, err := getInterfaceUsageBytes(snmp, ifIndex)
	if err != nil {
		return 0, err
	}
	secondGetTime := time.Now()

	// calcurate
	ifUsagePercent := calcInterfaceUsagePercent(firstUsageBytesMetric, secondUsageBytesMetric, firstGetTime.Second(), secondGetTime.Second(), linkCapBits)

	return ifUsagePercent, nil
}
