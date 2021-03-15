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
const MegaBitsToBits = 1000000
const ifHighSpeedOID = "1.3.6.1.2.1.31.1.1.1.15"
const ifHCInOctetsOID = "1.3.6.1.2.1.31.1.1.1.6"
const ifHCOutOctetsOID = "1.3.6.1.2.1.31.1.1.1.10"
const ifNumberOID = "1.3.6.1.2.1.2.1.0"
const ifDescrOID = "1.3.6.1.2.1.2.2.1.2"

type NetworkInterface struct {
	Sid           string
	NextSids      []string
	InterfaceName string
	LinkCap       int64
}

func GatherMetricsBySNMP(networkInterfaces []*NetworkInterface, interval int, srnodeAddr string, srnodePort uint16, snmpUser, snmpAuthPass, snmpPrivPass string, rateFlag bool) ([]*api.Node, error) {
	var eg errgroup.Group
	var nodes []*api.Node

	mutex := &sync.Mutex{}
	for _, ni := range networkInterfaces {
		ni := ni
		sc := newSNMPClient(srnodeAddr, srnodePort, snmpUser, snmpAuthPass, snmpPrivPass)
		eg.Go(func() error {
			return func(ni *NetworkInterface) error {
				ifIndex, err := getInterfaceIndexByName(sc, ni.InterfaceName)
				if err != nil {
					return err
				}

				usage, err := getInterfaceUsageBySNMP(sc, ifIndex, interval, ni.LinkCap, rateFlag)
				if err != nil {
					return err
				}

				// todo 欲しいのは空き帯域幅なのでとりあえずここで変換する
				// 		ただしusageという名前は違和感がある
				//		redisのデータ構造から見直さないとだめ
				if rateFlag == false {
					usage = float64(ni.LinkCap) - usage
				}

				for _, ns := range ni.NextSids {
					node := api.Node{
						SID:       ni.Sid,
						LinkCosts: []*api.LinkCost{NewLinkCost(ns, usage)},
					}
					mutex.Lock()
					nodes = append(nodes, &node)
					mutex.Unlock()
				}
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

func newSNMPClient(addr string, port uint16, user, authPass, privPass string) *gosnmp.GoSNMP {
	return &gosnmp.GoSNMP{
		Target:        addr,
		Port:          port,
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

	oids := []string{ifNumberOID}
	res, err := snmp.Get(oids)
	if err != nil {
		return 0, fmt.Errorf("failed get interface number from snmp agent: %v", err)
	}

	var maxIfIndex int
	for _, variable := range res.Variables {
		if variable.Type != gosnmp.Integer {
			return 0, fmt.Errorf("variable type is wrong correct %v, got %v", gosnmp.Integer, variable.Type)
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
			return 0, fmt.Errorf("variable type is wrong correct %v, got %v", gosnmp.Counter64, variable.Type)
		}
		totalBytes.Add(totalBytes, gosnmp.ToBigInt(variable.Value))
	}
	return totalBytes.Int64(), nil
}

func calcInterfaceUsage(firstBytes, secondBytes int64, duration float64, linkCapBits int64, rateFlag bool) (ifUsage float64) {
	traficBytesDiff := secondBytes - firstBytes
	if rateFlag {
		ifUsage = float64(traficBytesDiff) / (duration * float64(linkCapBits)) * BytesToBits * 100.0
	} else {
		ifUsage = float64(traficBytesDiff) / duration
	}
	if ifUsage < 0 {
		ifUsage = 0
	}
	return
}

func getInterfaceUsageBySNMP(snmp *gosnmp.GoSNMP, ifIndex, secInterval int, linkCap int64, rateFlag bool) (float64, error) {
	if linkCap <= 0 {
		var err error
		linkCap, err = getInterfaceCapacity(snmp, ifIndex)
		if err != nil {
			return 0, err
		}
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
	dur := secondGetTime.Sub(firstGetTime).Seconds()
	ifUsage := calcInterfaceUsage(firstUsageBytesMetric, secondUsageBytesMetric, dur, linkCap*MegaBitsToBits, rateFlag)

	return ifUsage, nil
}
