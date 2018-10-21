package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/spf13/viper"
	"github.com/vajafari/messagehub/cmd/server/internal/hub"
)

func main() {
	err := configViper()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	rand.Seed(time.Now().UTC().UnixNano())

	h := hub.NewEndpoint(getEndpointConf())
	h.Start()

	//fmt.Printf("Starting end point")
	//err = tcp.NewEndpoint(getEndpointConf(), nil).Start()
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	// fmt.Println("Endpoint stoped serving")
}

func configViper() error {
	// Init viper to load configs of server
	viper.SetConfigName("serverconfig") // no need to include file extension
	viper.AddConfigPath(".")            // set the path of your config file
	return viper.ReadInConfig()
}

func getEndpointConf() hub.EndpointConfing {
	return hub.EndpointConfing{
		Host:          viper.GetString("host"),
		Port:          viper.GetInt("port"),
		NetType:       viper.GetString("netType"),
		SendQueueSize: viper.GetInt("sendQueueSize"),
		ReadBufSize:   viper.GetInt("readBufSize"),
		WriteBufSize:  viper.GetInt("writeBufSize"),
		HubQueueSize:  viper.GetInt("hubQueueSize"),
	}
}
