package socket

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/pkg/errors"
)

const (
	readTimeout  = 2 * time.Hour
	writeTimeout = 120 * time.Second
)

// TCPSocket holds information a connection betwen client and server
// This class designed to be resuable across projeces
type TCPSocket struct {
	conn      *net.TCPConn    // TCP connection.
	id        uint64          // Assigned ID to current TCPSocket
	send      chan Frame      // Outgoing frames queue. We use a buffered channel of frames as thread-safe FIFO queue
	closeGoes chan bool       // This channel use to stop all go routines of TCPSocekt
	readChan  chan<- RData    // if successfull read happen signal send through this channle
	writeChan chan<- WData    // if successfull write happen signal send through this channle
	probChan  chan<- ProbData // if error occur signal send through this channle
	// This is map that specify how musch data is valid for ech type of message
	// We use this to prevent client from sending illogical data.
	// Each frame type (first byte of frame) has max lenght
	msgTypeLen   map[byte]uint32
	readBufSize  int
	writeBufSize int
}

//NewTCPSocket create tcpSocket object to hold client collection info
func NewTCPSocket(conn *net.TCPConn, id uint64, sendChanSize int, readBufSize int, writeBufSize int) *TCPSocket {
	s := TCPSocket{
		conn:         conn,
		send:         make(chan Frame, sendChanSize),
		closeGoes:    make(chan bool, 1),
		readBufSize:  readBufSize,
		writeBufSize: writeBufSize,
		id:           id,
	}

	log.Printf("TCPSocket, Createing new TCPSocket with ID: %d-readBuf:%d-writeBuf:%d-sendQueue:%d", s.id, readBufSize, writeBufSize, sendChanSize)

	return &s
}

//Start set channel to communication with socket manager
func (s *TCPSocket) Start(writeChan chan<- WData, readChan chan<- RData, probChan chan<- ProbData, msgTypeLen map[byte]uint32) {
	if s.writeChan != nil || s.readChan != nil || s.probChan != nil || s.msgTypeLen != nil {
		return
	}
	s.readChan = readChan
	s.writeChan = writeChan
	s.probChan = probChan
	s.msgTypeLen = msgTypeLen
	go s.reader()
	go s.writer()
	log.Printf("Start called for socket %d\n", s.id)
}

//Send Add packet to send queue
func (s *TCPSocket) Send(frm Frame) {
	s.send <- frm
	log.Printf("New frame is pushed to send queue for socket %d\n", s.id)
}

// Close tcpSocket and release all the resources
func (s *TCPSocket) Close() error {
	log.Printf("TcpSocket, CLOSE function for socket %d called\n", s.id)
	err := s.conn.Close()
	if err == nil {
		s.closeGoes <- true
		s.closeGoes <- true
		close(s.send)
		close(s.closeGoes)
		log.Printf("TcpSocket, socket %d closed\n", s.id)
	} else {
		log.Println(err)
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
	log.Printf("ITcpSocket, D ser for socket %d\n", s.id)
}

func getFramePrefix() []byte {
	return []byte{83, 79, 70, 10}
}

func (s *TCPSocket) writer() {
	log.Printf("TcpSocket, Go routine writer started for socket %d\n", s.id)
	prefix := getFramePrefix()
	bufio.NewWriterSize(s.conn, s.writeBufSize)
	buf := bufio.NewWriter(s.conn)
	for {
		//When we close 'send' channel, It still received buffered frames so
		//writer go routine continue to work. we use following select to make sure that
		//this go routine closes immediately when we call TCPSocket close method
		select {
		case <-s.closeGoes:
			log.Printf("TcpSocket, Close signal recieved by writer go routine of socket %d. Write routine terminated\n", s.id)
			return
		case frm := <-s.send:
			log.Printf("TcpSocket, Sending packet process start from tcpSocket %d\n", s.id)
			if frm == nil {
				continue
			}
			bb := frm.Serialize()
			if len(bb) == 0 {
				continue
			}
			bb = append(prefix, bb...)

			nn, err := s.writeWithRetry(buf, bb, writeTimeout)
			if err != nil {
				res := ProbData{
					Frm:      frm,
					SourceID: s.id,
					Err:      err,
				}
				s.probChan <- res
				if err == io.EOF {
					log.Printf("TCPSocket, EOF detected for socker %d. Write routine terminated\n", s.id)
					return
				}
				if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
					log.Printf("TCPSocket, TIMEOUT detected for socker %d. Write routine terminated\n", s.id)
					return
				}

			} else {
				res := WData{
					Frm:      frm,
					SourceID: s.id,
				}
				log.Printf("TcpSocket, SUCCESSFULL! Sent  message to tcpSocket %d. Message len was %d", s.id, nn)
				s.writeChan <- res
			}
		}
	}
}

func (s *TCPSocket) reader() {
	//We make a buffered read to reduce read syscalls.
	log.Printf("TcpSocket, Go routine reader started for socket %d\n", s.id)
	buf := bufio.NewReaderSize(s.conn, s.readBufSize)
	bb := make([]byte, s.readBufSize)
	for {
		select {
		case <-s.closeGoes:
			log.Printf("TcpSocket, Close signal recieved by reader go routine of socket %d. Write routine terminated\n.", s.id)
			return
		default:
		}
		s.conn.SetReadDeadline(time.Now().Add(readTimeout))
		n, err := buf.Read(bb)
		if err != nil {
			s.probChan <- ProbData{
				Err:      err,
				SourceID: s.ID(),
			}
			fmt.Printf("TcpSocket, Error On Read %s\n", err.Error())
			if err == io.EOF {
				log.Printf("TCPSocket, EOF detected for socker %d. Read routine terminated\n", s.id)
				return
			}
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				log.Printf("TCPSocket, TIMEOUT detected for socker %d. Read routine terminated\n", s.id)
				return
			}
			//break
		} else {
			fmt.Printf("TcpSocket, %d bytes recieved\n", n)
			fmt.Println(bb)
		}

	}

}

func (s *TCPSocket) writeWithRetry(buf *bufio.Writer, bb []byte, timeout time.Duration) (int, error) {
	s.conn.SetWriteDeadline(time.Now().Add(timeout))
	nn, err := buf.Write(bb)
	if err != nil {
		log.Printf("TcpSocket, Error on writing data to tcpSocket %d. Error Message is %s\n", s.id, err.Error())
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			s.conn.SetWriteDeadline(time.Now().Add(timeout))
			nn, err = buf.Write(bb)
		}
		if err != nil {
			err = errors.Wrapf(err, "TcpSocket, Error on REwrite data to tcpSocket %d. Error Message is %s", s.id, err.Error())
			return nn, err
		}
	}
	err = buf.Flush()
	if err != nil {
		log.Printf("TcpSocket, Error on flushing buffer message for tcpSocket %d. Error Message is %s\n", s.id, err.Error())
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			s.conn.SetWriteDeadline(time.Now().Add(timeout))
			err = buf.Flush()
			if err != nil {
				err = errors.Wrapf(err, "TcpSocket, Error on REflushing data to tcpSocket %d. Error Message is %s", s.id, err.Error())
			}
		}
	}
	return nn, err
}
