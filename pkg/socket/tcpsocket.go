package socket

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/pkg/errors"
)

const (
	//hdrTimeout = 2 * time.Second
	msgTimeout = 60 * time.Second
)

// TCPSocket holds information a connection betwen client and server
// This class designed to be resuable across projeces
type TCPSocket struct {
	conn      *net.TCPConn    // TCP connection.
	send      chan Frame      // Outgoing frames queue. We use a buffered channel of frames as thread-safe FIFO queue
	id        uint64          // Assigned ID to current TCPSocket
	readChan  chan<- RData    // if successfull read happen signal send through this channle
	writeChan chan<- WData    // if successfull write happen signal send through this channle
	probChan  chan<- ProbData // if error occur signal send through this channle
	closeGoes chan bool       // This channel use to stop all go routines of TCPSocekt
	// This is map that specify how musch data is valid for ech type of message
	// We use this to prevent client from sending illogical data.
	// Each frame type (first byte of frame) has max lenght
	msgTypeLen map[byte]uint32
}

//NewTCPSocket create tcpSocket object to hold client collection info
func NewTCPSocket(conn *net.TCPConn, sendBufSize int, id uint64, msgTypeLen map[byte]uint32, readChan chan<- RData, writeChan chan<- WData, probChan chan<- ProbData) *TCPSocket {
	s := TCPSocket{
		conn:       conn,
		send:       make(chan Frame, sendBufSize),
		id:         id,
		msgTypeLen: msgTypeLen,
		readChan:   readChan,
		writeChan:  writeChan,
		probChan:   probChan,
		closeGoes:  make(chan bool),
	}

	log.Printf("Createing new TCPSocket with ID: %d\n", s.id)
	go s.reader()
	go s.writer()
	return &s
}

//SetCommChannels set channel to communication with socket manager
func (s *TCPSocket) SetCommChannels(writeChan chan<- WData, readChan chan<- RData, probChan chan<- ProbData) {
	s.readChan = readChan
	s.writeChan = writeChan
	s.probChan = probChan
	log.Printf("SetCommChannels called for socket %d\n", s.id)
}

//Send Add packet to send queue
func (s *TCPSocket) Send(frm Frame) {
	s.send <- frm
	log.Printf("New frame is pushed to send queue for socket %d\n", s.id)
}

// Close tcpSocket and release all the resources
func (s *TCPSocket) Close() error {
	log.Printf("Closeing socket %d\n", s.id)
	err := s.conn.Close()
	if err != nil {
		s.closeGoes <- true
		s.closeGoes <- true
		close(s.send)
		log.Printf("socket %d\n closed\n", s.id)
	}
	return err
}

// ID return current channel id
func (s *TCPSocket) ID() uint64 {
	return s.id
}

// SetID return current channel id
func (s *TCPSocket) SetID(id uint64) {
	s.id = id
	log.Printf("ID ser for socket %d\n", s.id)
}

func getFramePrefix() []byte {
	return []byte{83, 79, 70, 10}
}

func (s *TCPSocket) writer() {
	log.Printf("Go routine writer started for socket %d\n", s.id)
	prefix := getFramePrefix()
	buf := bufio.NewWriter(s.conn)
	for frm := range s.send {
		//When we close 'send' channel, It still received buffered frames so
		//writer go routine continue to work. we use following select to make sure that
		//this go routine closes immediately when we call TCPSocket close method
		select {
		case <-s.closeGoes:
			log.Printf("Close signal recieved by writer go routine of socket %d\n", s.id)
			break
		default:
		}

		log.Printf("Sending packet process start from tcpSocket %d\n", s.id)
		bb := frm.Serialize()
		if len(bb) == 0 {
			continue
		}
		bb = append(prefix, bb...)

		nn, err := s.writeWithRetry(buf, bb, msgTimeout)
		if err != nil {
			log.Printf("Error on sending frame from socket %d. Error message %s", s.id, err.Error())
			res := ProbData{
				Frm:      frm,
				SourceID: s.id,
				Err:      err,
			}
			s.probChan <- res

		} else {
			res := WData{
				Frm:      frm,
				SourceID: s.id,
			}
			log.Printf("SUCCESSFULL! Sent  message to tcpSocket %d. Message len was %d", s.id, nn)
			s.writeChan <- res
		}
	}
	log.Printf("Go routine writer reach to end for socket %d\n", s.id)
}

func (s *TCPSocket) reader() {
	//We make a buffered read to reduce read syscalls.
	buf := bufio.NewReader(s.conn)

	for {
		log.Printf("Start receiving command from tcpSocket %d\n", s.id)
		bb := make([]byte, 10)
		n, err := buf.Read(bb)
		if err != nil {
			s.probChan <- ProbData{
				Err:      err,
				SourceID: s.ID(),
			}
			fmt.Printf("Error On Read %s\n", err.Error())
		} else {
			fmt.Printf("%d bytes recieved\n", n)
			fmt.Println(bb)
		}

	}
	log.Printf("Go routine reader reach to end for socket %d\n", s.id)
}

func (s *TCPSocket) writeWithRetry(buf *bufio.Writer, bb []byte, timeout time.Duration) (int, error) {
	s.conn.SetWriteDeadline(time.Now().Add(timeout))
	nn, err := buf.Write(bb)
	if err != nil {
		log.Printf("Error on writing data to tcpSocket %d. Error Message is %s\n", s.id, err.Error())
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			s.conn.SetWriteDeadline(time.Now().Add(timeout))
			nn, err = buf.Write(bb)
		}
		if err != nil {
			err = errors.Wrapf(err, "Error on REwrite data to tcpSocket %d. Error Message is %s", s.id, err.Error())
			return nn, err
		}
	}
	err = buf.Flush()
	if err != nil {
		log.Printf("Error on flushing buffer message for tcpSocket %d. Error Message is %s\n", s.id, err.Error())
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			s.conn.SetWriteDeadline(time.Now().Add(timeout))
			err = buf.Flush()
			if err != nil {
				err = errors.Wrapf(err, "Error on REflushing data to tcpSocket %d. Error Message is %s", s.id, err.Error())
			}
		}
	}
	return nn, err
}
