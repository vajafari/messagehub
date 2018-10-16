package hub

import (
	"log"
	"testing"
	"time"

	"github.com/vajafari/messagehub/pkg/message"

	"github.com/vajafari/messagehub/pkg/socket"
)

type packetMock struct {
	typ  byte
	data []byte
}

func (pkt packetMock) Type() byte {
	return pkt.typ
}
func (pkt packetMock) Data() ([]byte, error) {
	return pkt.data, nil
}

type socketMock struct {
	id         uint64
	readChan   chan<- socket.RData
	writeChan  chan<- socket.WData
	probChan   chan<- socket.ProbData
	msgTypeLen map[byte]uint32
	packets    []socket.Packet
}

func (s *socketMock) Start(writeChan chan<- socket.WData, readChan chan<- socket.RData, probChan chan<- socket.ProbData, msgTypeLen map[byte]uint32) {
	s.readChan = readChan
	s.writeChan = writeChan
	s.probChan = probChan
	s.msgTypeLen = msgTypeLen
	s.packets = make([]socket.Packet, 0)
}

func (s *socketMock) Close() error {
	return nil
}
func (s *socketMock) ID() uint64 {
	return s.id
}
func (s *socketMock) SetID(id uint64) {
	s.id = id
}
func (s *socketMock) Send(pkt socket.Packet) {
	s.packets = append(s.packets, pkt)
}

func (s *socketMock) clearPackets() {
	s.packets = make([]socket.Packet, 0)
}

func (s *socketMock) simulateReadData(bb []byte) {
	if len(bb) > 1 {
		s.readChan <- socket.RData{
			Pkt: packetMock{
				data: bb[1:],
				typ:  bb[0],
			},
			SourceID: s.ID(),
		}
	} else {
		s.readChan <- socket.RData{
			Pkt: packetMock{
				data: nil,
				typ:  bb[0],
			},
			SourceID: s.ID(),
		}
	}
}

func TestReadHandler(t *testing.T) {
	h := NewHub(100)
	sMock1 := socketMock{id: 1}
	h.Add(&sMock1)
	sMock2 := socketMock{id: 2}
	h.Add(&sMock2)
	sMock3 := socketMock{id: 3}
	h.Add(&sMock3)
	sMock4 := socketMock{id: 4}
	h.Add(&sMock4)
	// 4 socket exist in hub now

	// Simulate message Id
	sMock1.simulateReadData([]byte{byte(message.IDMgsCode)})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) != 1 {
		log.Fatal("Error on response to IDRequestMsg")
	}
	if len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		log.Fatal("Error on response to IDRequestMsg. Id message response sent to wrong clients")
	}

	if sMock1.packets[0].Type() != byte(message.IDMgsCode) {
		log.Fatal("Error on response to IDRequestMsg. Response message code is not valid")
	}

	dataSMock, err := sMock1.packets[0].Data()
	if err != nil {
		log.Fatal("Error on response to IDRequestMsg. Cannot deserialize message on client")
	}
	idRespMsg, err := message.DeserializeIDRes(dataSMock)
	if err != nil {
		log.Fatal("Error on response to IDRequestMsg. Cannot deserialize message on client")
	}
	if idRespMsg.ID != sMock1.id {
		log.Fatal("Error on response to IDRequestMsg. Wrong Id sent to client")
	}
	sMock1.clearPackets()
	sMock1.simulateReadData([]byte{byte(message.IDMgsCode), 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) != 1 {
		log.Fatal("Error on response to IDRequestMsg")
	}

	// Test List Message
	sMock1.clearPackets()
	//Request from unidentified socket
	sMock1.simulateReadData([]byte{byte(message.ListMgsCode)})
	time.Sleep(20 * time.Millisecond)

	if len(sMock1.packets) > 0 || len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		log.Fatal("Error on response to ListRequestMsg. Generate list response based on request from unindentified socket")
	}

	//Request recieved from identified socket, but no identifed channel exist
	h.socketRepo[1].IsIdentified = true
	sMock1.simulateReadData([]byte{byte(message.ListMgsCode)})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) != 1 {
		log.Fatal("Error on response to List request")
	}
	if sMock1.packets[0].Type() != byte(message.ListMgsCode) {
		log.Fatal("Error on response to ListRequestMsg. Response message code is not valid")
	}
	if len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		log.Fatal("Error on response to ListRequestMsg. List message response sent to wrong clients")
	}
	dataSMock, err = sMock1.packets[0].Data()
	if err != nil {
		log.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}
	listRespMsg, err := message.DeserializeListRes(dataSMock)
	if err != nil {
		log.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}
	if !message.ChkListResponseMsgEq(listRespMsg, message.ListResponseMsg{}) {
		log.Fatal("Error on response to ListRequestMsg. Wrong List message recieved on client")
	}

	sMock1.clearPackets()
	h.socketRepo[3].IsIdentified = true
	sMock1.simulateReadData([]byte{byte(message.ListMgsCode)})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) != 1 {
		log.Fatal("Error on response to List request")
	}
	if sMock1.packets[0].Type() != byte(message.ListMgsCode) {
		log.Fatal("Error on response to ListRequestMsg. Response message code is not valid")
	}
	if len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		log.Fatal("Error on response to ListRequestMsg. List message response sent to wrong clients")
	}
	dataSMock, err = sMock1.packets[0].Data()
	if err != nil {
		log.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}
	listRespMsg, err = message.DeserializeListRes(dataSMock)
	if err != nil {
		log.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}
	if !message.ChkListResponseMsgEq(listRespMsg, message.ListResponseMsg{IDs: []uint64{3}}) {
		log.Fatal("Error on response to ListRequestMsg. Wrong List message recieved on client")
	}

	// Test Relay Request
	sMock1.clearPackets()
	sMock2.clearPackets()
	sMock3.clearPackets()
	sMock4.clearPackets()
	h.socketRepo[1].IsIdentified = false
	h.socketRepo[2].IsIdentified = false
	h.socketRepo[3].IsIdentified = false
	h.socketRepo[4].IsIdentified = false

	sMock1.simulateReadData([]byte{byte(message.RelayMgsCode), 2, 2, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7})
	time.Sleep(20 * time.Millisecond)

	if len(sMock1.packets) > 0 || len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		log.Fatal("Error on response to RelayRequestMsg. Generate relay response based on request from unindentified socket")
	}

	h.socketRepo[1].IsIdentified = true
	sMock1.clearPackets()
	sMock1.simulateReadData([]byte{byte(message.RelayMgsCode), 2, 2, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) > 0 || len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		log.Fatal("Error on response to ListRequestMsg. List message response sent to wrong clients")
	}

	sMock1.clearPackets()
	sMock2.clearPackets()
	sMock3.clearPackets()
	sMock4.clearPackets()
	h.socketRepo[3].IsIdentified = true
	h.socketRepo[4].IsIdentified = true
	sMock1.simulateReadData([]byte{byte(message.RelayMgsCode), 3, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7})

	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) > 0 || len(sMock2.packets) > 0 {
		log.Fatal("Error on response to RelayRequestMsg. Id message response sent to wrong clients")
	}
	if len(sMock3.packets) == 0 || len(sMock4.packets) == 0 {
		log.Fatal("Error on response to RelayRequestMsg")
	}
	if sMock3.packets[0].Type() != byte(message.RelayMgsCode) {
		log.Fatal("Error on response to RelayRequestMsg. Response message code is not valid")
	}
	if sMock4.packets[0].Type() != byte(message.RelayMgsCode) {
		log.Fatal("Error on response to RelayRequestMsg. Response message code is not valid")
	}
	dataSMock, err = sMock3.packets[0].Data()
	if err != nil {
		log.Fatal("Error on response to RelayRequestMsg. Cannot deserialize message on client")
	}
	relayRespMsg, err := message.DeserializeRelayRes(dataSMock)
	if err != nil {
		log.Fatal("Error on response to RelayRequestMsg. Cannot deserialize message on client")
	}
	if !message.ChkRelayResponseMsgEq(relayRespMsg, message.RelayResponseMsg{SenderID: 1, Body: []byte{1, 2, 3, 4, 5, 6, 7}}) {
		log.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}
	dataSMock, err = sMock4.packets[0].Data()
	if err != nil {
		log.Fatal("Error on response to RelayRequestMsg. Cannot deserialize message on client")
	}
	relayRespMsg, err = message.DeserializeRelayRes(dataSMock)
	if err != nil {
		log.Fatal("Error on response to RelayRequestMsg. Cannot deserialize message on client")
	}
	if !message.ChkRelayResponseMsgEq(relayRespMsg, message.RelayResponseMsg{SenderID: 1, Body: []byte{1, 2, 3, 4, 5, 6, 7}}) {
		log.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}

	//Test incorect message format
	sMock1.clearPackets()
	sMock2.clearPackets()
	sMock3.clearPackets()
	sMock4.clearPackets()
	sMock1.simulateReadData([]byte{byte(message.RelayMgsCode), 10, 2, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) > 0 || len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		log.Fatal("Error on response to RelayRequestMsg. Shouldnot reponse to invalid message")
	}
}
