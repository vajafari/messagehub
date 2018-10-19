package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vajafari/messagehub/pkg/proxy"
	"github.com/vajafari/messagehub/pkg/socket"
)

func main() {
	err := configViper()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	clientConfig := getClientConf()
	prx, err := connect(clientConfig)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	r := bufio.NewReader(os.Stdin)
	for {
		//time.Sleep(2 * time.Second)
		fmt.Println("Please enter yout command:")
		fmt.Println("[1]- Send ID Request")
		fmt.Println("[2]- Send List Request")
		fmt.Println("[3]- Send Relay Request")
		fmt.Println("[4]- Exit")
		cmd, err := scanInput(r)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		switch cmd {
		case 1:
			prx.SendID()
		case 2:
			prx.SendList()
		case 3:
			fmt.Println("How many client Id?")
			ids := make([]uint64, 0)
			cliCnt, err := scanInput(r)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			for i := 0; i < int(cliCnt); i++ {
				fmt.Printf("Enter clinet Id %d\n", i+1)
				cliID, err := scanInput(r)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
				ids = append(ids, cliID)
			}
			fmt.Println("How many bytes for body?")
			bytesCnt, err := scanInput(r)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			bb := make([]byte, bytesCnt)
			for i := 0; i < int(cliCnt); i++ {
				bb[i] = byte(i % 256)
			}
			prx.SendRelay(ids, bb)
		case 4:
			return
		default:
			fmt.Println("Invalid command")
		}
	}

}

func connect(clientConfig ClientConfig) (*proxy.Proxy, error) {
	fmt.Printf("Connecting to %s ...\n", clientConfig.GetHostAddress())
	d := net.Dialer{Timeout: time.Second * time.Duration(clientConfig.DailTimeout)}
	conn, err := d.Dial(clientConfig.NetType, clientConfig.GetHostAddress())
	if err != nil {
		return nil, err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil, errors.New("Connection is not TCP")
	}
	skt := socket.NewTCPSocket(tcpConn, 0, clientConfig.SendQueueSize, clientConfig.ReadBufSize, clientConfig.WriteBufSize)
	prx := proxy.NewProxy(clientConfig.ProxyQueueSize)
	err = prx.SetSocket(skt)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected!!")
	return prx, nil
}

func scanInput(r *bufio.Reader) (uint64, error) {
	text, err := r.ReadString('\n')
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}
	text = strings.TrimSuffix(text, "\r\n")
	return strconv.ParseUint(text, 0, 64)
}
