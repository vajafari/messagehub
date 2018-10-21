package hub

import (
	"errors"
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
	msgTypeLen map[byte]int
	packets    []socket.Packet
	closed     bool
}

func (s *socketMock) Start(writeChan chan<- socket.WData, readChan chan<- socket.RData, probChan chan<- socket.ProbData, msgTypeLen map[byte]int) {
	s.readChan = readChan
	s.writeChan = writeChan
	s.probChan = probChan
	s.msgTypeLen = msgTypeLen
	s.packets = make([]socket.Packet, 0)
}

func (s *socketMock) Close() error {
	s.closed = true
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

func (s *socketMock) simulateProbData(pkt socket.Packet, err error) {
	s.probChan <- socket.ProbData{
		Pkt:      pkt,
		SourceID: s.ID(),
		Err:      err,
	}
}

func (s *socketMock) simulateReadData(pkt socket.Packet) {
	s.readChan <- socket.RData{
		Pkt:      pkt,
		SourceID: s.ID(),
	}
}

func (s *socketMock) simulateReadDataByte(bb []byte) {
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

func (s *socketMock) simulateWriteData(pkt socket.Packet) {
	s.writeChan <- socket.WData{
		Pkt:      pkt,
		SourceID: s.id,
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
	//sMock1.simulateReadData(  []byte{byte(message.IDMgsCode)})
	sMock1.simulateReadData(message.IDRequestMsg{})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) != 1 {
		t.Fatal("Error on response to IDRequestMsg")
	}
	if len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		t.Fatal("Error on response to IDRequestMsg. Id message response sent to wrong clients")
	}

	if sMock1.packets[0].Type() != byte(message.IDMgsCode) {
		t.Fatal("Error on response to IDRequestMsg. Response message code is not valid")
	}

	dataSMock, err := sMock1.packets[0].Data()
	if err != nil {
		t.Fatal("Error on response to IDRequestMsg. Cannot deserialize message on client")
	}
	idRespMsg, err := message.DeserializeIDRes(dataSMock)
	if err != nil {
		t.Fatal("Error on response to IDRequestMsg. Cannot deserialize message on client")
	}
	if idRespMsg.ID != sMock1.id {
		t.Fatal("Error on response to IDRequestMsg. Wrong Id sent to client")
	}
	sMock1.clearPackets()
	//sMock1.simulateReadData([]byte{byte(message.IDMgsCode), 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	sMock1.simulateReadData(message.IDRequestMsg{})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) != 1 {
		t.Fatal("Error on response to IDRequestMsg")
	}

	// Test List Message
	sMock1.clearPackets()
	//Request from unidentified socket
	//sMock1.simulateReadData([]byte{byte(message.ListMgsCode)})
	sMock1.simulateReadData(message.ListRequestMsg{})
	time.Sleep(20 * time.Millisecond)

	if len(sMock1.packets) > 0 || len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		t.Fatal("Error on response to ListRequestMsg. Generate list response based on request from unindentified socket")
	}

	//Request recieved from identified socket, but no identifed channel exist
	h.sktRepo[1].IsIdentified = true
	//sMock1.simulateReadData([]byte{byte(message.ListMgsCode)})
	sMock1.simulateReadData(message.ListRequestMsg{})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) != 1 {
		t.Fatal("Error on response to List request")
	}
	if sMock1.packets[0].Type() != byte(message.ListMgsCode) {
		t.Fatal("Error on response to ListRequestMsg. Response message code is not valid")
	}
	if len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		t.Fatal("Error on response to ListRequestMsg. List message response sent to wrong clients")
	}
	dataSMock, err = sMock1.packets[0].Data()
	if err != nil {
		t.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}
	listRespMsg, err := message.DeserializeListRes(dataSMock)
	if err != nil {
		t.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}
	if !message.ChkListResponseMsgEq(listRespMsg, message.ListResponseMsg{}) {
		t.Fatal("Error on response to ListRequestMsg. Wrong List message recieved on client")
	}

	sMock1.clearPackets()
	h.sktRepo[3].IsIdentified = true
	sMock1.simulateReadData(message.ListRequestMsg{})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) != 1 {
		t.Fatal("Error on response to List request")
	}
	if sMock1.packets[0].Type() != byte(message.ListMgsCode) {
		t.Fatal("Error on response to ListRequestMsg. Response message code is not valid")
	}
	if len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		t.Fatal("Error on response to ListRequestMsg. List message response sent to wrong clients")
	}
	dataSMock, err = sMock1.packets[0].Data()
	if err != nil {
		t.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}
	listRespMsg, err = message.DeserializeListRes(dataSMock)
	if err != nil {
		t.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}
	if !message.ChkListResponseMsgEq(listRespMsg, message.ListResponseMsg{IDs: []uint64{3}}) {
		t.Fatal("Error on response to ListRequestMsg. Wrong List message recieved on client")
	}

	// Test Relay Request
	sMock1.clearPackets()
	sMock2.clearPackets()
	sMock3.clearPackets()
	sMock4.clearPackets()
	h.sktRepo[1].IsIdentified = false
	h.sktRepo[2].IsIdentified = false
	h.sktRepo[3].IsIdentified = false
	h.sktRepo[4].IsIdentified = false

	sMock1.simulateReadData(message.RelayRequestMsg{IDs: []uint64{2, 4}, Body: []byte{1, 2, 3, 4, 5, 6, 7}})
	time.Sleep(20 * time.Millisecond)

	if len(sMock1.packets) > 0 || len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		t.Fatal("Error on response to RelayRequestMsg. Generate relay response based on request from unindentified socket")
	}

	h.sktRepo[1].IsIdentified = true
	sMock1.clearPackets()
	sMock1.simulateReadData(message.RelayRequestMsg{IDs: []uint64{2, 4}, Body: []byte{1, 2, 3, 4, 5, 6, 7}})
	//sMock1.simulateReadData([]byte{byte(message.RelayMgsCode), 2, 2, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) > 0 || len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		t.Fatal("Error on response to ListRequestMsg. List message response sent to wrong clients")
	}

	sMock1.clearPackets()
	sMock2.clearPackets()
	sMock3.clearPackets()
	sMock4.clearPackets()
	h.sktRepo[3].IsIdentified = true
	h.sktRepo[4].IsIdentified = true
	sMock1.simulateReadData(message.RelayRequestMsg{IDs: []uint64{2, 3, 4}, Body: []byte{1, 2, 3, 4, 5, 6, 7}})
	//sMock1.simulateReadData([]byte{byte(message.RelayMgsCode), 3, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7})

	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) > 0 || len(sMock2.packets) > 0 {
		t.Fatal("Error on response to RelayRequestMsg. Id message response sent to wrong clients")
	}
	if len(sMock3.packets) == 0 || len(sMock4.packets) == 0 {
		t.Fatal("Error on response to RelayRequestMsg")
	}
	if sMock3.packets[0].Type() != byte(message.RelayMgsCode) {
		t.Fatal("Error on response to RelayRequestMsg. Response message code is not valid")
	}
	if sMock4.packets[0].Type() != byte(message.RelayMgsCode) {
		t.Fatal("Error on response to RelayRequestMsg. Response message code is not valid")
	}
	dataSMock, err = sMock3.packets[0].Data()
	if err != nil {
		t.Fatal("Error on response to RelayRequestMsg. Cannot deserialize message on client")
	}
	relayRespMsg, err := message.DeserializeRelayRes(dataSMock)
	if err != nil {
		t.Fatal("Error on response to RelayRequestMsg. Cannot deserialize message on client")
	}
	if !message.ChkRelayResponseMsgEq(relayRespMsg, message.RelayResponseMsg{SenderID: 1, Body: []byte{1, 2, 3, 4, 5, 6, 7}}) {
		t.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}
	dataSMock, err = sMock4.packets[0].Data()
	if err != nil {
		t.Fatal("Error on response to RelayRequestMsg. Cannot deserialize message on client")
	}
	relayRespMsg, err = message.DeserializeRelayRes(dataSMock)
	if err != nil {
		t.Fatal("Error on response to RelayRequestMsg. Cannot deserialize message on client")
	}
	if !message.ChkRelayResponseMsgEq(relayRespMsg, message.RelayResponseMsg{SenderID: 1, Body: []byte{1, 2, 3, 4, 5, 6, 7}}) {
		t.Fatal("Error on response to ListRequestMsg. Cannot deserialize message on client")
	}

	//Test incorect message format
	sMock1.clearPackets()
	sMock2.clearPackets()
	sMock3.clearPackets()
	sMock4.clearPackets()
	sMock1.simulateReadDataByte([]byte{byte(message.RelayMgsCode), 10, 2, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7})

	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) > 0 || len(sMock2.packets) > 0 || len(sMock3.packets) > 0 || len(sMock4.packets) > 0 {
		t.Fatal("Error on response to RelayRequestMsg. Shouldnot reponse to invalid message")
	}
}

func TestAdd(t *testing.T) {
	h := NewHub(100)
	if len(h.sktRepo) > 0 {
		t.Fatalf("New hub cannot have socket. Socket len %d", len(h.sktRepo))
	}
	sMock1 := socketMock{id: 1}
	h.Add(&sMock1)
	if len(h.sktRepo) != 1 {
		t.Fatalf("Error on add new socket. Actual len %d expected len %d", len(h.sktRepo), 1)
	}
	if h.sktRepo[1].IsIdentified {
		t.Error("New socket cannon be identified")
	}
	//sMock1.simulateReadData([]byte{byte(message.IDMgsCode)})
	sMock1.simulateReadData(message.IDRequestMsg{})
	time.Sleep(20 * time.Millisecond)
	if len(sMock1.packets) != 1 {
		t.Error("Channels not set to socket correctly")
	}
}

func TestWriteHandler(t *testing.T) {
	h := NewHub(100)
	sMock1 := socketMock{id: 1}
	h.Add(&sMock1)
	sMock1.simulateWriteData(message.IDRequestMsg{})
	time.Sleep(20 * time.Millisecond)
	if !h.sktRepo[1].IsIdentified {
		t.Fatal("Send to socket not reported to hub")
	}
}

func TestCloseSocket(t *testing.T) {
	h := NewHub(100)
	sMock1 := socketMock{id: 1}
	h.Add(&sMock1)
	sMock2 := socketMock{id: 2}
	h.Add(&sMock2)
	sMock3 := socketMock{id: 3}
	h.Add(&sMock3)

	sMock1.simulateProbData(message.IDRequestMsg{}, errors.New("Error on send"))
	time.Sleep(20 * time.Millisecond)
	if _, ok := h.sktRepo[1]; ok {
		t.Fatalf("Error does not close channel")
	}
	if len(h.sktRepo) != 2 {
		t.Fatalf("Error does not close channel")
	}
	if !sMock1.closed {
		t.Fatalf("Close methods of socket not called")
	}
}
