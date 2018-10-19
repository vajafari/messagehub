package proxy

import (
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

func TestSetSocket(t *testing.T) {
	prx := NewProxy(100)
	sMock1 := socketMock{}
	sMock2 := socketMock{}

	err := prx.SetSocket(nil)
	if err == nil {
		t.Fatal("Nil socket set to proxt")
	}
	err = prx.SetSocket(&sMock1)
	if err != nil {
		t.Fatalf("Error on set socket to proxy. Error message %s", err.Error())
	}

	err = prx.SetSocket(&sMock2)
	if err == nil {
		t.Fatal("Override current proxy")
	}
}

func TestCloseSocket(t *testing.T) {
	prx := NewProxy(100)
	sMock1 := socketMock{}
	sMock2 := socketMock{}

	prx.SetSocket(&sMock1)
	err := prx.CloseSocket()
	if err != nil {
		t.Fatalf("Error on close socket. Error message %s", err.Error())
	}
	if prx.skt != nil {
		t.Fatalf("Socket not removed from proxy")
	}
	if !sMock1.closed {
		t.Fatalf("Socket close method not called")
	}

	err = prx.SetSocket(&sMock2)
	if err != nil {
		t.Fatal("Proxy socket closed before, but cannot set new socket")
	}
}

func TestIdentification(t *testing.T) {
	prx := NewProxy(100)
	sMock1 := socketMock{}
	prx.SetSocket(&sMock1)

	sMock1.simulateReadData(message.IDResponseMsg{ID: 12})
	time.Sleep(20 * time.Millisecond)

	if prx.skt.ID() != 12 {
		t.Fatal("Id not set correctly")
	}

	sMock1.simulateReadData(message.IDResponseMsg{ID: 15})
	time.Sleep(20 * time.Millisecond)
	if prx.skt.ID() != 12 {
		t.Fatal("Cannot reset ID of socket")
	}
}

// SendIDReq send ID message to server via socket
func TestSendIDReq(t *testing.T) {
	prx := NewProxy(100)
	sMock1 := socketMock{}
	err := prx.SendID()
	if err == nil {
		t.Fatal("Cannot send id request when no socket set to proxy")
	}
	prx.SetSocket(&sMock1)
	err = prx.SendID()
	if len(sMock1.packets) != 1 {
		t.Fatal("Id request not sent to socket")
	}
	sMock1.clearPackets()
	sMock1.id = 12
	err = prx.SendID()
	if err == nil {
		t.Fatal("Send id request againt for identified socket")
	}
	if len(sMock1.packets) > 0 {
		t.Fatal("Send id request againt for identified socket")
	}

}

// SendListReq send list message to server via socket
func TestSendListReq(t *testing.T) {
	prx := NewProxy(100)
	sMock1 := socketMock{}
	err := prx.SendList()
	if err == nil {
		t.Fatal("Cannot send list request when no socket set to proxy")
	}

	prx.SetSocket(&sMock1)
	err = prx.SendList()
	if err == nil {
		t.Fatal("Cannot send list request when socket not identified")
	}
	if len(sMock1.packets) > 0 {
		t.Fatal("Cannot send list request when socket not identified")
	}
	sMock1.id = 12
	err = prx.SendList()
	if err != nil {
		t.Fatal("Error on send list request")
	}
	if len(sMock1.packets) != 1 {
		t.Fatal("List request not sent to socket")
	}
}

func TestSendRelayRequest(t *testing.T) {
	prx := NewProxy(100)
	sMock1 := socketMock{}

	bbOk := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	bbNotOk := make([]byte, message.RelayMaxBodySize+100)
	IdsOk := []uint64{250, 260, 270}
	IdsNotOk := make([]uint64, 300)
	for i := 0; i < 300; i++ {
		IdsNotOk[i] = uint64(i + 1)
	}
	for i := 0; i < message.RelayMaxBodySize+100; i++ {
		bbNotOk[i] = 1
	}

	err := prx.SendRelay(IdsOk, bbOk)
	if err == nil {
		t.Fatal("No socket set for this proxy")
	}

	prx.SetSocket(&sMock1)
	err = prx.SendRelay(IdsOk, bbOk)
	if err == nil {
		t.Fatal("Cannot send relay request when socket not identified")
	}
	if len(sMock1.packets) > 0 {
		t.Fatal("Cannot send relay request when socket not identified")
	}
	sMock1.id = 12

	err = prx.SendRelay(IdsOk, bbNotOk)
	if err == nil {
		t.Fatal("Cannot send invalid relay message")
	}
	err = prx.SendRelay(IdsNotOk, bbOk)
	if err == nil {
		t.Fatal("Cannot send invalid relay message")
	}
	err = prx.SendRelay(IdsNotOk, bbNotOk)
	if err == nil {
		t.Fatal("Cannot send invalid relay message")
	}
	err = prx.SendRelay(IdsOk, bbOk)
	if err != nil {
		t.Fatal("Valid relay not sent to proxy")
	}
	if len(sMock1.packets) != 1 {
		t.Fatal("Cannot send list request when socket not identified")
	}

}
