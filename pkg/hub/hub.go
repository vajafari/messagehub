package hub

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/vajafari/messagehub/pkg/message"

	"github.com/vajafari/messagehub/pkg/socket"
)

const (
	maxIDMsgLen    uint32 = 0       // Max length for id message in hub
	maxListMsgLen  uint32 = 0       // Max length for list message in hub
	maxRelayMsgLen uint32 = 1050617 // Max length for list message (1024 * 1024) + (255 * 8) + 1
)

// Hub is connection manager of specific server
// Design hub as separate module improve the scalability of the system
// we can create different hubs for each type of messages
// and assign them to the different end point
type Hub struct {
	socketRepo   map[uint64]*socketInfo
	mutx         sync.RWMutex
	readChan     chan socket.RData
	writeChan    chan socket.WData
	probChan     chan socket.ProbData
	msgTypeLen   map[byte]uint32
	readBufSize  int
	writeBufSize int
}

// NewHub Create new instance and initilize properties of hub struct
func NewHub(chanBuffer int) *Hub {

	hub := Hub{
		socketRepo: make(map[uint64]*socketInfo),
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
func (h *Hub) Add(skt socket.Socket) error {
	if skt == nil {
		return errors.New("Hub cannot accept nil sockets")
	}
	if skt.ID() == 0 {
		return errors.New("Socket must have id")
	}

	info := socketInfo{
		Skt:          skt,
		IsIdentified: false,
	}
	h.mutx.Lock()
	defer h.mutx.Unlock()
	if _, ok := h.socketRepo[skt.ID()]; ok {
		return errors.New("Socket with same ID already exist in hub. Please release all the resources of socket")
	}
	h.socketRepo[skt.ID()] = &info
	skt.Start(h.writeChan, h.readChan, h.probChan, h.msgTypeLen)
	fmt.Printf("Hub, Item add to map. Map len %d\n", len(h.socketRepo))
	return nil
}

func (h *Hub) readHandler() {
	log.Println("Hub, Starting hub READER handler Go routine")
	for rData := range h.readChan {
		switch rData.MsgType {
		case byte(message.IDMgsCode):
			h.handleIDReq(rData)
		case byte(message.ListMgsCode):
			h.handleListReq(rData)
		case byte(message.RelayMgsCode):
			h.handleRelayReq(rData)
		default:
			fmt.Printf("Hub, Invalid message recieved from scoket %d\n", rData.SourceID)
		}
	}
	log.Println("Hub, Stop hub READER handler Go routine")
}

func (h *Hub) handleIDReq(reqData socket.RData) {
	h.mutx.RLock()
	defer h.mutx.RUnlock()
	log.Printf("Hub, ID request received from socket %d\n", reqData.SourceID)
	if sktInfo, ok := h.socketRepo[reqData.SourceID]; ok {
		fmt.Printf("Hub, Id response to socket %d push in send queue\n", reqData.SourceID)
		sktInfo.Skt.Send(message.IDResponseMsg{ID: sktInfo.Skt.ID()})
	}
}

func (h *Hub) handleListReq(reqData socket.RData) {
	h.mutx.RLock()
	defer h.mutx.RUnlock()
	log.Printf("Hub, List request received from socket %d\n", reqData.SourceID)
	if sktInfo, ok := h.socketRepo[reqData.SourceID]; ok {
		if !sktInfo.IsIdentified {
			fmt.Printf("Hub, Reject list message from unidentified socket %d\n", reqData.SourceID)
			return
		}
		//TODO: Benchmark this
		i := 0
		for k, v := range h.socketRepo {
			if k != reqData.SourceID && v.IsIdentified {
				i++
			}
		}
		connList := make([]uint64, i)
		i = 0
		for k, v := range h.socketRepo {
			if k != reqData.SourceID && v.IsIdentified {
				connList[i] = v.Skt.ID()
				i++
			}
		}
		fmt.Printf("Hub, List response to socket %d push in send queue. IDs:{%v}\n", reqData.SourceID, reqData.SourceID)
		sktInfo.Skt.Send(message.ListResponseMsg{IDs: connList})
	}
}

func (h *Hub) handleRelayReq(reqData socket.RData) {
	h.mutx.RLock()
	defer h.mutx.RUnlock()
	log.Printf("Hub, relay message received from socket %d\n", reqData.SourceID)
	if sktInfo, ok := h.socketRepo[reqData.SourceID]; ok {
		if !sktInfo.IsIdentified {
			fmt.Printf("Hub, reject relay message from unidentified socket %d\n", reqData.SourceID)
			return
		}
	}

	msg, err := message.DeserializeRelayReq(reqData.Data)
	if err == nil {
		rspMsg := message.RelayResponseMsg{
			Data:     msg.Data,
			SenderID: reqData.SourceID,
		}
		for _, id := range msg.IDs {
			if sktInfo, ok := h.socketRepo[id]; ok {
				if sktInfo.IsIdentified {
					fmt.Printf("Hub, relay response to socket %d push in send queue\n", id)
					sktInfo.Skt.Send(rspMsg)
				}
			}
		}
	} else {
		log.Printf("Hub, error on DeserializeRelayReq of relay message from socket {%d}\n", reqData.SourceID)
	}
}

func (h *Hub) writeHandler() {
	log.Println("Hub, starting WRITER handler Go routine")
	for wData := range h.writeChan {

		if wData.Frm.Type() == byte(message.IDMgsCode) {
			h.mutx.Lock()
			defer h.mutx.Unlock()
			if sktInfo, ok := h.socketRepo[wData.SourceID]; ok {
				sktInfo.IsIdentified = true
			}
		}
	}
	log.Println("Hub, stoping hub WRITER handler Go routine")
}

func (h *Hub) probHandler() {
	log.Println("Hub, starting hub PROB handler Go routine")
	for sig := range h.probChan {
		fmt.Println("Hub, prob Recived.")
		h.closeSocket(sig.SourceID)
	}
	log.Println("Hub, stoping hub PROB handler Go routine")
}

// closeSocket find specific socket by id and close it
func (h *Hub) closeSocket(id uint64) {
	h.mutx.Lock()
	defer h.mutx.Unlock()
	fmt.Printf("Hub, close socket %d\n", id)
	if sktInfo, ok := h.socketRepo[id]; ok {
		fmt.Printf("Hub, Socket found in hub. Call close function of %d\n", sktInfo.Skt.ID())
		err := sktInfo.Skt.Close()
		if err != nil {
			log.Printf("Hub, Error on closeing channel %d. Error messag is %s\n", sktInfo.Skt.ID(), err.Error())
		}
		delete(h.socketRepo, id)
	} else {
		fmt.Printf("Hub, Error On find socker %d in socket map. socker map count is %d\n", id, len(h.socketRepo))
	}
}

// Connected sockets info
type socketInfo struct {
	Skt          socket.Socket
	IsIdentified bool
}
