package hub

import (
	"errors"
	"fmt"
	"sync"

	"github.com/vajafari/messagehub/pkg/message"

	"github.com/vajafari/messagehub/pkg/socket"
)

const (
	maxIDMsgLen   int = 0 // Max length for id message in hub
	maxListMsgLen int = 0 // Max length for list message in hub
	// Max length for relay message (1024 * 1024) + (255 * 8) + 1
	maxRelayMsgLen int = int(message.RelayMaxBodySize + (message.RelayMaxReciverCount * 8) + 1)
)

// Hub is connection manager of specific server
// Design hub as separate module improve the scalability of the system
// we can create different hubs for each type of messages
// and assign them to the different end point
type Hub struct {
	sktRepo    map[uint64]*socketInfo
	mutx       sync.RWMutex
	readChan   chan socket.RData
	writeChan  chan socket.WData
	probChan   chan socket.ProbData
	msgTypeLen map[byte]int
}

// NewHub Create new instance and initilize properties of hub struct
func NewHub(queueSize int) *Hub {

	hub := Hub{
		sktRepo:    make(map[uint64]*socketInfo),
		readChan:   make(chan socket.RData, queueSize),
		writeChan:  make(chan socket.WData, queueSize),
		probChan:   make(chan socket.ProbData, queueSize),
		msgTypeLen: make(map[byte]int),
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
	if _, ok := h.sktRepo[skt.ID()]; ok {
		return errors.New("Socket with same ID already exist in hub. Please release all the resources of socket")
	}

	h.sktRepo[skt.ID()] = &info
	skt.Start(h.writeChan, h.readChan, h.probChan, h.msgTypeLen)
	fmt.Printf("Hub, Item add to map. Map len %d\n", len(h.sktRepo))
	return nil
}

func (h *Hub) readHandler() {
	fmt.Println("Hub, Starting hub READER handler Go routine")
	for rData := range h.readChan {
		switch rData.Pkt.Type() {
		case byte(message.IDMgsCode):
			go h.handleIDReq(rData)
		case byte(message.ListMgsCode):
			go h.handleListReq(rData)
		case byte(message.RelayMgsCode):
			go h.handleRelayReq(rData)
		default:
			fmt.Printf("Hub, Invalid message recieved from scoket %d\n", rData.SourceID)
		}
	}
	fmt.Println("Hub, Stop hub READER handler Go routine")
}

func (h *Hub) handleIDReq(reqData socket.RData) {
	h.mutx.RLock()
	defer h.mutx.RUnlock()
	fmt.Printf("Hub, ID request received from socket %d\n", reqData.SourceID)
	if sktInfo, ok := h.sktRepo[reqData.SourceID]; ok {
		fmt.Printf("Hub, Id response to socket %d push in send queue\n", reqData.SourceID)
		sktInfo.Skt.Send(message.IDResponseMsg{ID: sktInfo.Skt.ID()})
	} else {
		fmt.Printf("Hub, Socket ID %d not found in hub", reqData.SourceID)
	}
}

func (h *Hub) handleListReq(reqData socket.RData) {
	h.mutx.RLock()
	defer h.mutx.RUnlock()
	fmt.Printf("Hub, List request received from socket %d\n", reqData.SourceID)
	if sktInfo, ok := h.sktRepo[reqData.SourceID]; ok {
		if !sktInfo.IsIdentified {
			fmt.Printf("Hub, Reject list message from unidentified socket %d\n", reqData.SourceID)
			return
		}
		i := 0
		for k, v := range h.sktRepo {
			if k != reqData.SourceID && v.IsIdentified {
				i++
			}
		}
		connList := make([]uint64, i)
		i = 0
		for k, v := range h.sktRepo {
			if k != reqData.SourceID && v.IsIdentified {
				connList[i] = v.Skt.ID()
				i++
			}
		}
		fmt.Printf("Hub, List response to socket %d push in send queue. IDs:{%v}\n", reqData.SourceID, reqData.SourceID)
		if len(connList) > message.ListMaxItems {
			sktInfo.Skt.Send(message.ListResponseMsg{IDs: connList[0:message.ListMaxItems]})
		} else {
			sktInfo.Skt.Send(message.ListResponseMsg{IDs: connList})
		}

	} else {
		fmt.Printf("Hub, Socket ID %d not found in hub", reqData.SourceID)
	}
}

func (h *Hub) handleRelayReq(reqData socket.RData) {
	h.mutx.RLock()
	defer h.mutx.RUnlock()
	fmt.Printf("Hub, relay message received from socket %d\n", reqData.SourceID)
	if sktInfo, ok := h.sktRepo[reqData.SourceID]; ok {
		if !sktInfo.IsIdentified {
			fmt.Printf("Hub, reject relay message from unidentified socket %d\n", reqData.SourceID)
			return
		}
	} else {
		fmt.Printf("Hub, Socket ID %d not found in hub", reqData.SourceID)
		return
	}
	data, err := reqData.Pkt.Data()
	if err != nil {
		fmt.Printf("Hub, error on DeserializeRelayReq of relay message from socket {%d}\n", reqData.SourceID)
		return
	}
	msg, err := message.DeserializeRelayReq(data)
	if err == nil {
		rspMsg := message.RelayResponseMsg{
			Body:     msg.Body,
			SenderID: reqData.SourceID,
		}
		for _, id := range msg.IDs {
			if sktInfo, ok := h.sktRepo[id]; ok {
				if sktInfo.IsIdentified {
					fmt.Printf("Hub, relay response to socket %d push in send queue\n", id)
					sktInfo.Skt.Send(rspMsg)
				}
			}
		}
	} else {
		fmt.Printf("Hub, error on DeserializeRelayReq of relay message from socket {%d}\n", reqData.SourceID)
	}
}

func (h *Hub) writeHandler() {
	fmt.Println("Hub, starting WRITER handler Go routine")
	for wData := range h.writeChan {

		if wData.Pkt.Type() == byte(message.IDMgsCode) {
			h.mutx.Lock()
			if sktInfo, ok := h.sktRepo[wData.SourceID]; ok {
				sktInfo.IsIdentified = true
			}
			h.mutx.Unlock()
		}
	}
	fmt.Println("Hub, stoping hub WRITER handler Go routine")
}

func (h *Hub) probHandler() {
	fmt.Println("Hub, starting hub PROB handler Go routine")
	for sig := range h.probChan {
		fmt.Println("Hub, prob Recived.")
		h.closeSocket(sig.SourceID)
	}
	fmt.Println("Hub, stoping hub PROB handler Go routine")
}

// closeSocket find specific socket by id and close it
func (h *Hub) closeSocket(id uint64) {
	h.mutx.Lock()
	defer h.mutx.Unlock()
	fmt.Printf("Hub, close socket %d\n", id)
	if sktInfo, ok := h.sktRepo[id]; ok {
		fmt.Printf("Hub, Socket found in hub. Call close function of %d\n", sktInfo.Skt.ID())
		err := sktInfo.Skt.Close()
		if err != nil {
			fmt.Printf("Hub, Error on closeing channel %d. Error messag is %s\n", sktInfo.Skt.ID(), err.Error())
		}
		delete(h.sktRepo, id)
		fmt.Printf("HUB, Remove channel number %d, HUB LEN %d\n", id, len(h.sktRepo))
	} else {
		fmt.Printf("Hub, Error On find socker %d in socket map. socker map count is %d\n", id, len(h.sktRepo))
	}
}

// Connected sockets info
type socketInfo struct {
	Skt          socket.Socket
	IsIdentified bool
}
