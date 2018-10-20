package proxy

import (
	"errors"
	"fmt"
	"sync"

	"github.com/vajafari/messagehub/pkg/message"
	"github.com/vajafari/messagehub/pkg/socket"
)

var (
	// ErrNotConnected happn when no socket set to proxy
	ErrNotConnected = errors.New("No socket set for this proxy")
	// ErrNotIdentified happen when try to send list or relay message to the hub
	ErrNotIdentified = errors.New("This socket is not identified")
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

// NewProxy Create a new instance and initialize properties of the proxy struct
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

// SetSocket process send and receive data
func (prx *Proxy) SetSocket(skt socket.Socket) error {
	if skt == nil {
		return ErrNotConnected
	}
	if prx.skt != nil {
		return errors.New("A socket already set for proxy")
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
		fmt.Printf("Proxy, Error on closing socket. Error message %s\n", err.Error())
		return err
	}
	prx.skt = nil
	fmt.Println("Proxy, Socket closed!")
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
	fmt.Println("Proxy, Id message pushed in socket send queue")
	return nil
}

// SendList send list message to hub via socket
func (prx *Proxy) SendList() error {
	prx.mutx.RLock()
	defer prx.mutx.RUnlock()
	if prx.skt == nil {
		return ErrNotConnected
	}
	if prx.skt.ID() == 0 {
		return ErrNotIdentified
	}
	prx.skt.Send(message.ListRequestMsg{})
	fmt.Println("Proxy, List message pushed in socket send queue")
	return nil
}

// SendRelay send relay message to hub via socket
func (prx *Proxy) SendRelay(ids []uint64, bb []byte) error {
	prx.mutx.RLock()
	defer prx.mutx.RUnlock()
	if prx.skt == nil {
		return ErrNotConnected
	}
	if prx.skt.ID() == 0 {
		return ErrNotIdentified
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
	fmt.Println("Proxy, Relay message pushed in socket send queue")
	return nil
}

func (prx *Proxy) readHandler() {
	for rData := range prx.readChan {
		switch rData.Pkt.Type() {
		case byte(message.IDMgsCode):
			prx.handleIDReq(rData)
		case byte(message.ListMgsCode):
			prx.handleListReq(rData)
		case byte(message.RelayMgsCode):
			prx.handleRelayReq(rData)
		default:
			fmt.Println("Proxy, Invalid message received from scoket")
		}
	}
}

func (prx *Proxy) handleIDReq(reqData socket.RData) {
	bb, err := reqData.Pkt.Data()
	if err != nil {
		fmt.Println("Proxy, Error on retrieving id message")
		return
	}
	msg, err := message.DeserializeIDRes(bb)
	if err != nil {
		fmt.Println("Proxy, Error on deserializing id message")
		return
	}
	prx.mutx.Lock()
	defer prx.mutx.Unlock()
	if prx.skt.ID() == 0 {
		prx.skt.SetID(msg.ID)
		fmt.Printf("Id response received. Client id: %d\n", msg.ID)
	} else if prx.skt.ID() != msg.ID {
		fmt.Println("Another id assigned to client before")
	}

}

func (prx *Proxy) handleListReq(reqData socket.RData) {

	bb, err := reqData.Pkt.Data()
	if err != nil {
		fmt.Println("Proxy, Error on retrieving list message")
		return
	}
	msg, err := message.DeserializeListRes(bb)
	if err != nil {
		fmt.Println("Proxy, Error on deserializing list message")
		return
	}
	fmt.Printf("List response received: %v\n", msg.IDs)
}

func (prx *Proxy) handleRelayReq(reqData socket.RData) {
	bb, err := reqData.Pkt.Data()
	if err != nil {
		fmt.Println("Proxy, Error on retrieving relay message")
		return
	}
	msg, err := message.DeserializeRelayRes(bb)
	if err != nil {
		fmt.Println("Proxy, Error on deserializing relay message")
		return
	}
	fmt.Printf("Relay response received. Message length is %d, sender id is %d\n", len(msg.Body), msg.SenderID)
}

func (prx *Proxy) writeHandler() {
	for {
		<-prx.writeChan
	}
}

func (prx *Proxy) probHandler() {
	for sig := range prx.probChan {
		fmt.Printf("Hub, Problem Recived. Error message is %s", sig.Err.Error())
		prx.CloseSocket()
	}
}
