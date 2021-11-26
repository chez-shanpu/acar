package controlplane

import "github.com/spf13/viper"

const (
	DataplaneAddr = "dp-addr"
	DataplaneCert = "dp-cert"
	DataplaneTLS  = "dp-tls"
	ServerAddr    = "addr"
	TLS           = "tls"
	TLSCert       = "cert"
	TLSKey        = "key"
)

type DaemonConfig struct {
	DataplaneAddr string
	DataplaneCert string
	DataplaneTLS  bool
	ServerAddr    string
	TLS           bool
	TLSCert       string
	TLSKey        string
}

var Config = &DaemonConfig{}

func (c *DaemonConfig) Populate() {
	c.DataplaneAddr = viper.GetString(DataplaneAddr)
	c.DataplaneCert = viper.GetString(DataplaneCert)
	c.DataplaneTLS = viper.GetBool(DataplaneTLS)
	c.ServerAddr = viper.GetString(ServerAddr)
	c.TLS = viper.GetBool(TLS)
	c.TLSCert = viper.GetString(TLSCert)
	c.TLSKey = viper.GetString(TLSKey)
}
