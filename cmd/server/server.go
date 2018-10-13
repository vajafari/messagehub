package main

import (
	"github.com/spf13/viper"
	"github.com/vajafari/messagehub/pkg/hub"
)

func main() {

	//var bbbb []byte
	//fmt.Println(len(bbbb))
	// bb := make([]byte, 8)
	// binary.LittleEndian.PutUint64(bb, uint64(4739))
	// slc := []uint64{3, 27, 492, 4587, 87345, 159743, 1468743, 22446078, 374557900, 3459980326, 348529584035, 3849560104835, 39475690128456}

	// for _, n := range slc {
	// 	binary.LittleEndian.PutUint64(bb, uint64(n))
	// 	fmt.Print(bb[0])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[1])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[2])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[3])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[4])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[5])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[6])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[7])
	// 	fmt.Print(", ")
	// }
	// err := configViper()
	// if err != nil {
	// 	log.Println(err.Error())

	// 	return
	// }

	// rand.Seed(time.Now().UTC().UnixNano())

	// log.Printf("Starting end point")
	// endpoint := hub.NewEndpoint(getEndpointConf())
	// err = endpoint.Start()
	// if err != nil {
	// 	log.Println(err.Error())
	// }
	// log.Println("Endpoint stoped serving")

}

func configViper() error {
	// Init viper to load configs of server
	viper.SetConfigName("serverconfig") // no need to include file extension
	viper.AddConfigPath(".")            // set the path of your config file
	return viper.ReadInConfig()
}

func getEndpointConf() hub.EndpointConfing {
	return hub.EndpointConfing{
		Host:        viper.GetString("host"),
		Port:        viper.GetInt("port"),
		NetType:     viper.GetString("netType"),
		SendBufSize: viper.GetInt("sendBufSize"),
	}
}
