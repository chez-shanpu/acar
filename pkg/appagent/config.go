package appagent

import "github.com/spf13/viper"

const (
	AppName          = "app"
	ControlplaneAddr = "cp-addr"
	ControlplaneCert = "cp-cert"
	ControlplaneTLS  = "cp-tls"
	DepSID           = "dep-sid"
	DstAddr          = "dst-addr"
	DstSID           = "dst-sid"
	Interval         = "interval"
	Lazy             = "lazy"
	Metrics          = "metrics"
	MonitoringAddr   = "mnt-addr"
	MonitoringCert   = "mnt-cert"
	MonitoringTLS    = "mnt-tls"
	Require          = "require"
	SrcAddr          = "src-addr"
)

type DaemonConfig struct {
	AppName          string
	ControlplaneAddr string
	ControlplaneCert string
	ControlplaneTLS  bool
	DepSIDs          []string
	DstAddr          string
	DstSID           string
	Interval         int
	Lazy             float64
	MetricsType      string
	MonitoringAddr   string
	MonitoringCert   string
	MonitoringTLS    bool
	RequireValue     float64
	SrcAddr          string
}

var Config = &DaemonConfig{}

func (c *DaemonConfig) Populate() {
	c.AppName = viper.GetString(AppName)
	c.ControlplaneAddr = viper.GetString(ControlplaneAddr)
	c.ControlplaneCert = viper.GetString(ControlplaneCert)
	c.ControlplaneTLS = viper.GetBool(ControlplaneTLS)
	c.DepSIDs = viper.GetStringSlice(DepSID)
	c.DstAddr = viper.GetString(DstAddr)
	c.DstSID = viper.GetString(DstSID)
	c.Interval = viper.GetInt(Interval)
	c.Lazy = viper.GetFloat64(Lazy)
	c.MetricsType = viper.GetString(Metrics)
	c.MonitoringAddr = viper.GetString(MonitoringAddr)
	c.MonitoringCert = viper.GetString(MonitoringCert)
	c.MonitoringTLS = viper.GetBool(MonitoringTLS)
	c.RequireValue = viper.GetFloat64(Require)
	c.SrcAddr = viper.GetString(SrcAddr)
}
