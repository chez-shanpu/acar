package srnode

import "github.com/spf13/viper"

const (
	Addr               = "addr"
	ConcurrentNum      = "concurrent"
	ConcurrentInterval = "concurrent-interval"
	Interval           = "interval"
	MonitoringAddr     = "mnt-addr"
	Port               = "port"
	SNMPAuthPass       = "auth-pass"
	SNMPPrivPass       = "priv-pass"
	SNMPUser           = "snmp-user"
	TLS                = "tls"
	TLSCert            = "cert"
)

type DaemonConfig struct {
	Addr               string
	ConcurrentNum      int
	ConcurrentInterval float64
	Interval           int
	MonitoringAddr     string
	Port               int
	SNMPAuthPass       string
	SNMPPrivPass       string
	SNMPUser           string
	TLS                bool
	TLSCert            string
	NetworkInterfaces  []*NetworkInterface
}

var Config = &DaemonConfig{}

func (c *DaemonConfig) Populate() {
	c.Addr = viper.GetString(Addr)
	c.ConcurrentNum = viper.GetInt(ConcurrentNum)
	c.ConcurrentInterval = viper.GetFloat64(ConcurrentInterval)
	c.Interval = viper.GetInt(Interval)
	c.MonitoringAddr = viper.GetString(MonitoringAddr)
	c.Port = viper.GetInt(Port)
	c.SNMPAuthPass = viper.GetString(SNMPAuthPass)
	c.SNMPPrivPass = viper.GetString(SNMPPrivPass)
	c.SNMPUser = viper.GetString(SNMPUser)
	c.TLS = viper.GetBool(TLS)
	c.TLSCert = viper.GetString(TLSCert)
	_ = viper.Unmarshal(c)
}
