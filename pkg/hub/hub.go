package hub

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"

	"github.com/vajafari/messagehub/pkg/message"

	"github.com/vajafari/messagehub/pkg/socket"
)

const (
	maxIDMsgLen    uint32 = 8       // Max length for id message
	maxListMsgLen  uint32 = 1048576 // Max length for list message (131072 client)
	maxRelayMsgLen uint32 = 1050617 // Max length for list message (1024 * 1024) + (255 * 8) + 1
)

// Hub is connection manager of specific server
// Design hub as separate module improve the scalability of the system
// we can create different hubs for each type of messages
// and assign them to the different end point

type Hub struct {
	// sync.Map Has much better performance for stable key cache system with concurrent
	// loops over map specially when CPU cores cores increase
	socketRepo sync.Map
	readChan   chan socket.RData
	writeChan  chan socket.WData
	probChan   chan socket.ProbData
	msgTypeLen map[byte]uint32
}

// NewHub Create new instance and initilize properties of hub struct
func NewHub(chanBuffer int) *Hub {

	hub := Hub{
		readChan:   make(chan socket.RData, chanBuffer),
		writeChan:  make(chan socket.WData, chanBuffer),
		probChan:   make(chan socket.ProbData, chanBuffer),
		msgTypeLen: make(map[byte]uint32),
	}
	hub.msgTypeLen[byte(message.IDMgsCode)] = maxIDMsgLen
	hub.msgTypeLen[byte(message.ListMgsCode)] = maxListMsgLen
	hub.msgTypeLen[byte(message.RelayMgsCode)] = maxRelayMsgLen

	go hub.probHandler()
	go hub.readHandler()
	go hub.writeHandler()
	return &hub
}

// Add new connection to socket pool
func (h *Hub) Add(conn *net.TCPConn, sendBufSize int) {
	var skt socket.Socket
	skt = socket.NewTCPSocket(conn, sendBufSize, rand.Uint64(), h.msgTypeLen, h.readChan, h.writeChan, h.probChan)
	info := socketInfo{
		Skt:          skt,
		IsIdentified: false,
	}
	h.socketRepo.Store(skt.ID(), &info)
}

func (h *Hub) readHandler() {
	log.Println("Starting ")
	for frm := range s.send {

	}
}

func (h *Hub) writeHandler() {
	fmt.Println("Run Prob handler")
	for sig := range h.probChan {
		fmt.Println("Rpob Recived.")
		h.closeSocket(sig.SourceID)
	}

}

func (h *Hub) probHandler() {
	fmt.Println("Run Prob handler")
	for sig := range h.probChan {
		fmt.Println("Rpob Recived.")
		h.closeSocket(sig.SourceID)
	}

}

// func (h *Hub) handleClose() {
// 	for id := range h.closeChan {
// 		h.
// 	}
// }

// func (h *Hub) handleWrite() {
// 	for data := range h.writeChan {

// 	}
// }

func (h *Hub) handleRead() {
	// for rwData := range h.readChan {
	// 	typ := rwData.Frm.GetHeader().Type
	// 	switch typ {
	// 	case 1:
	// 		go processIdMsg(rwData.SourceID)
	// 		// case 2:
	// 		// 	go processListMsg(rwData.SourceID)
	// 		// case 3:
	// 		// 	go processListMsg(wrData)
	// 	}

	// }
}

// func processIdMsg(id uint64) {
// 	if sktInfo, ok := h.socketRepo.Load(id); ok {
// 		var frm socket.Frame
// 		sktInfo.(socketInfo).Skt.Send(a)
// 	}
// }

// closeSocket find specific socket by id and close it
func (h *Hub) closeSocket(id uint64) {
	fmt.Printf("Closing socket %d", id)
	if sktInfo, ok := h.socketRepo.Load(id); ok {
		fmt.Println("Socker find")
		fmt.Println("Socker find %d", sktInfo.(socketInfo).Skt.ID())
		err := sktInfo.(socketInfo).Skt.Close()
		if err != nil {
			log.Printf("Error on closeing channel %d. Error messag is %s", sktInfo.(socketInfo).Skt.ID(), err.Error())
		}
	}
}

// Connected sockets infor
type socketInfo struct {
	Skt          socket.Socket
	IsIdentified bool
}
