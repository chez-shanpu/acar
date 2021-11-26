package monitoring

import "github.com/spf13/viper"

const (
	RedisAddr  = "redis-addr"
	RedisDB    = "redis-db"
	RedisPass  = "redis-pass"
	ServerAddr = "addr"
	TLS        = "tls"
	TLSCert    = "cert"
	TLSKey     = "key"
)

var Config = &DaemonConfig{}

type DaemonConfig struct {
	RedisAddr  string
	RedisDB    int
	RedisPass  string
	ServerAddr string
	TLS        bool
	TLSCert    string
	TLSKey     string
}

func (c *DaemonConfig) Populate() {
	c.RedisAddr = viper.GetString(RedisAddr)
	c.RedisDB = viper.GetInt(RedisDB)
	c.RedisPass = viper.GetString(RedisPass)
	c.ServerAddr = viper.GetString(ServerAddr)
	c.TLS = viper.GetBool(TLS)
	c.TLSCert = viper.GetString(TLSCert)
	c.TLSKey = viper.GetString(TLSKey)
}
