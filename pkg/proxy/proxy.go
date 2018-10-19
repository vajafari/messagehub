package proxy

import (
	"errors"
	"fmt"
	"sync"

	"github.com/vajafari/messagehub/pkg/message"
	"github.com/vajafari/messagehub/pkg/socket"
)

var (
	// ErrNotConnected happned when no socket set to proxy
	ErrNotConnected = errors.New("No socket set for this proxy")
)

const (
	maxIDMsgLen   int = 8                        // Max length for id message in cli
	maxListMsgLen int = message.ListMaxItems * 8 // Max message size
	// Max length for relay message: 1024 * 1024 bytes for body and 8 bytes for sender Id
	maxRelayMsgLen int = message.RelayMaxBodySize + 8
)

// Proxy is clinet side socket manager
type Proxy struct {
	skt        socket.Socket
	readChan   chan socket.RData
	writeChan  chan socket.WData
	probChan   chan socket.ProbData
	msgTypeLen map[byte]int
	mutx       sync.RWMutex
}

// NewProxy Create new instance and initilize properties of proxy struct
func NewProxy(queueSize int) *Proxy {
	prx := Proxy{
		readChan:   make(chan socket.RData, queueSize),
		writeChan:  make(chan socket.WData, queueSize),
		probChan:   make(chan socket.ProbData, queueSize),
		msgTypeLen: make(map[byte]int),
	}
	prx.msgTypeLen[byte(message.IDMgsCode)] = maxIDMsgLen
	prx.msgTypeLen[byte(message.ListMgsCode)] = maxListMsgLen
	prx.msgTypeLen[byte(message.RelayMgsCode)] = maxRelayMsgLen

	go prx.probHandler()
	go prx.readHandler()
	go prx.writeHandler()

	return &prx
}

// SetSocket process send and recieve data
func (prx *Proxy) SetSocket(skt socket.Socket) error {
	if skt == nil {
		return ErrNotConnected
	}
	if prx.skt != nil {
		return errors.New("A socket already set for proxy. Please close it first")
	}
	prx.mutx.Lock()
	defer prx.mutx.Unlock()
	prx.skt = skt
	prx.skt.Start(prx.writeChan, prx.readChan, prx.probChan, prx.msgTypeLen)
	return nil
}

// CloseSocket close current socket of proxy
func (prx *Proxy) CloseSocket() error {
	if prx.skt == nil {
		return ErrNotConnected
	}
	prx.mutx.Lock()
	defer prx.mutx.Unlock()
	err := prx.skt.Close()
	if err != nil {
		fmt.Printf("Error on socket closing process. Error message %s\n", err.Error())
		return err
	}
	prx.skt = nil
	return nil
}

// SendID send ID message to server via socket
func (prx *Proxy) SendID() error {
	prx.mutx.RLock()
	defer prx.mutx.RUnlock()
	if prx.skt == nil {
		return ErrNotConnected
	}
	if prx.skt.ID() > 0 {
		return errors.New("Id set to socket before")
	}
	prx.skt.Send(message.IDRequestMsg{})
	return nil
}

// SendList send list message to server via socket
func (prx *Proxy) SendList() error {
	prx.mutx.RLock()
	defer prx.mutx.RUnlock()
	if prx.skt == nil {
		return ErrNotConnected
	}
	if prx.skt.ID() == 0 {
		return errors.New("You are not identified. Send ID request first")
	}
	prx.skt.Send(message.ListRequestMsg{})
	return nil
}

// SendRelay send relay message to server via socket
func (prx *Proxy) SendRelay(ids []uint64, bb []byte) error {
	prx.mutx.RLock()
	defer prx.mutx.RUnlock()
	if prx.skt == nil {
		return ErrNotConnected
	}
	if prx.skt.ID() == 0 {
		return errors.New("You are not identified. Send ID request first")
	}
	if len(ids) > message.RelayMaxReciverCount || len(ids) == 0 {
		return errors.New("Recievers count is not valid")
	}
	if len(bb) > message.RelayMaxBodySize || len(bb) == 0 {
		return errors.New("Data len is not valid")
	}
	msg := message.RelayRequestMsg{
		Body: bb,
		IDs:  ids,
	}
	prx.skt.Send(msg)
	return nil
}

func (prx *Proxy) readHandler() {
	fmt.Println("Proxy, Starting READER handler Go routine")
	for rData := range prx.readChan {
		switch rData.Pkt.Type() {
		case byte(message.IDMgsCode):
			prx.handleIDReq(rData)
		case byte(message.ListMgsCode):
			prx.handleListReq(rData)
		case byte(message.RelayMgsCode):
			prx.handleRelayReq(rData)
		default:
			fmt.Printf("Proxy, Invalid message recieved from scoket %d\n", rData.SourceID)
		}
	}
	fmt.Println("Proxy, Stop proxy READER handler Go routine")
}

func (prx *Proxy) handleIDReq(reqData socket.RData) {
	bb, err := reqData.Pkt.Data()
	if err != nil {
		fmt.Println("Error on Deserialize ID response")
	}
	msg, err := message.DeserializeIDRes(bb)
	if err != nil {
		fmt.Println("Error on Deserialize ID response")
	}
	prx.mutx.RLock()
	defer prx.mutx.RUnlock()
	if prx.skt.ID() == 0 {
		prx.skt.SetID(msg.ID)
	} else {
		fmt.Println("Cannot reset ID")
	}
	fmt.Printf("ID response recieved. Socket ID is %d\n", msg.ID)
}

func (prx *Proxy) handleListReq(reqData socket.RData) {

	bb, err := reqData.Pkt.Data()
	if err != nil {
		fmt.Println("Error on Deserialize ID response")
	}
	msg, err := message.DeserializeListRes(bb)
	if err != nil {
		fmt.Println("Error on Deserialize ID response")
	}
	fmt.Printf("List response recieved. Socket ID is %v\n", msg.IDs)
}

func (prx *Proxy) handleRelayReq(reqData socket.RData) {
	bb, err := reqData.Pkt.Data()
	if err != nil {
		fmt.Println("Error on Deserialize ID response")
	}
	msg, err := message.DeserializeRelayRes(bb)
	if err != nil {
		fmt.Println("Error on Deserialize ID response")
	}
	fmt.Printf("Relay response recieved. Message length is %d, sender ID is %d\n", len(msg.Body), msg.SenderID)
}

func (prx *Proxy) writeHandler() {
	fmt.Println("Proxy, starting WRITER handler Go routine")
	for wData := range prx.writeChan {
		fmt.Printf("Proxy, Messge write in socket. type % d\n", wData.Pkt.Type())
	}
	fmt.Println("Proxy, stoping proxy WRITER handler Go routine")
}

func (prx *Proxy) probHandler() {
	fmt.Println("Proxy, starting proxy PROB handler Go routine")
	for sig := range prx.probChan {
		fmt.Printf("Proxy, Prob recived. error message is %s\n", sig.Err.Error())
		prx.CloseSocket()
	}
	fmt.Println("Proxy, stoping proxy PROB handler Go routine")
}
