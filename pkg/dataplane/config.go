package dataplane

import "github.com/spf13/viper"

const (
	Device     = "device"
	ServerAddr = "addr"
	TLS        = "tls"
	TLSCert    = "cert"
	TLSKey     = "key"
)

type DaemonConfig struct {
	Device     string
	ServerAddr string
	TLS        bool
	TLSCert    string
	TLSKey     string
}

var Config = &DaemonConfig{}

func (c *DaemonConfig) Populate() {
	c.Device = viper.GetString(Device)
	c.ServerAddr = viper.GetString(ServerAddr)
	c.TLS = viper.GetBool(TLS)
	c.TLSCert = viper.GetString(TLSCert)
	c.TLSKey = viper.GetString(TLSKey)
}
