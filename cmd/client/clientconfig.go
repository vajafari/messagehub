package main

import (
	"strconv"

	"github.com/spf13/viper"
)

// ClientConfig contain dynamic configurarion of client
type ClientConfig struct {
	Host           string
	Port           int
	NetType        string
	SendQueueSize  int
	ReadBufSize    int
	WriteBufSize   int
	ProxyQueueSize int
	DailTimeout    int
}

func configViper() error {
	// Init viper to load configs of server
	viper.SetConfigName("clientconfig") // no need to include file extension
	viper.AddConfigPath(".")            // set the path of your config file
	return viper.ReadInConfig()
}

func getClientConf() ClientConfig {
	return ClientConfig{
		Host:           viper.GetString("host"),
		Port:           viper.GetInt("port"),
		NetType:        viper.GetString("netType"),
		SendQueueSize:  viper.GetInt("sendQueueSize"),
		ReadBufSize:    viper.GetInt("readBufSize"),
		WriteBufSize:   viper.GetInt("writeBufSize"),
		ProxyQueueSize: viper.GetInt("proxyQueueSize"),
		DailTimeout:    viper.GetInt("dailTimeout"),
	}
}

// GetHostAddress Apprend host address and port number together and return  full address of site
func (conf *ClientConfig) GetHostAddress() string {
	return conf.Host + ":" + strconv.Itoa(conf.Port)
}
