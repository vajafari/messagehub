package socket

import (
	"bufio"
	"encoding/binary"
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
	prefixLen    = 7
)

var (
	packetPrefix = []byte{83, 79, 70, 83, 79, 70, 10}
)

// TCPSocket holds information a connection betwen client and server
// This class designed to be resuable across projeces
type TCPSocket struct {
	conn      *net.TCPConn    // TCP connection.
	id        uint64          // Assigned ID to current TCPSocket
	send      chan Packet     // Outgoing packets queue. We use a buffered channel of packets as thread-safe FIFO queue
	closeGoes chan bool       // This channel use to stop all go routines of TCPSocekt
	readChan  chan<- RData    // if successfull read happen signal send through this channle
	writeChan chan<- WData    // if successfull write happen signal send through this channle
	probChan  chan<- ProbData // if error occur signal send through this channle
	// This is map that specify how musch data is valid for ech type of message
	// We use this to prevent client from sending illogical data.
	// Each packet type (first byte of packet) has max lenght
	msgTypeLen   map[byte]int
	readBufSize  int
	writeBufSize int
}

//NewTCPSocket create tcpSocket object to hold client collection info
func NewTCPSocket(conn *net.TCPConn, id uint64, sendChanSize int, readBufSize int, writeBufSize int) *TCPSocket {
	s := TCPSocket{
		conn:         conn,
		send:         make(chan Packet, sendChanSize),
		closeGoes:    make(chan bool, 1),
		readBufSize:  readBufSize,
		writeBufSize: writeBufSize,
		id:           id,
	}

	log.Printf("TCPSocket, Createing new TCPSocket with ID: %d-readBuf:%d-writeBuf:%d-sendQueue:%d", s.id, readBufSize, writeBufSize, sendChanSize)

	return &s
}

//Start set channel to communication with socket manager
func (s *TCPSocket) Start(writeChan chan<- WData, readChan chan<- RData, probChan chan<- ProbData, msgTypeLen map[byte]int) {
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
func (s *TCPSocket) Send(pkt Packet) {
	s.send <- pkt
	log.Printf("New packet is pushed to send queue for socket %d\n", s.id)
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

func (s *TCPSocket) writer() {
	log.Printf("TcpSocket, Go routine writer started for socket %d\n", s.id)
	bufio.NewWriterSize(s.conn, s.writeBufSize)
	buf := bufio.NewWriter(s.conn)
	for {
		//When we close 'send' channel, It still received buffered packets so
		//writer go routine continue to work. we use following select to make sure that
		//this go routine closes immediately when we call TCPSocket close method
		select {
		case <-s.closeGoes:
			log.Printf("TcpSocket, Close signal recieved by writer go routine of socket %d. Write routine terminated\n", s.id)
			return
		case pkt := <-s.send:
			log.Printf("TcpSocket, Sending packet process start from tcpSocket %d\n", s.id)
			if pkt == nil {
				continue
			}
			// Prepare data for sending on wire!!!
			bb, err := pkt.Data()
			if err == nil {
				log.Printf("Error on get data from packet in socket %d\n", s.id)
			}

			pktBytes := make([]byte, prefixLen+HeaderLen+len(bb))
			copy(pktBytes, packetPrefix)     //Prefix
			pktBytes[prefixLen] = pkt.Type() //Type
			lenBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(lenBytes, uint32(len(bb)))
			copy(pktBytes[prefixLen+1:], lenBytes) //Lent
			copy(pktBytes[prefixLen+HeaderLen:], bb)
			nn, err := s.writeWithRetry(buf, bb, writeTimeout)
			if err != nil {
				res := ProbData{
					Pkt:      pkt,
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
					Pkt:      pkt,
					SourceID: s.id,
				}
				log.Printf("TcpSocket, SUCCESSFULL! Sent  message to tcpSocket %d. Message len was %d", s.id, nn)
				s.writeChan <- res
			}
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

func (s *TCPSocket) reader() {
	log.Printf("TcpSocket, Go routine reader started for socket %d\n", s.id)
	bb := make([]byte, s.readBufSize)
	buf := bufio.NewReaderSize(s.conn, s.readBufSize)
	pi := packetInspector{}
	pi.resetVariables()
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
			if n == 0 {
				continue
			}
			fmt.Printf("TcpSocket, %d bytes recieved in socket\n", n)
			packets := pi.inspect(bb[0:n], s.msgTypeLen)
			if len(packets) > 0 {
				for _, pkt := range packets {
					s.readChan <- RData{
						Pkt:      pkt,
						SourceID: s.id,
					}
				}
			}
		}
	}
}

type packetInspector struct {
	completeFindPrefix bool
	partialFindPrefix  bool
	headerVerified     bool
	prevPrefixCnt      int
	lastIndexPrefix    int
	currentPkgLen      int
	curPkgHeader       []byte
	curPkg             []byte
}

func (pi *packetInspector) resetVariables() {
	pi.completeFindPrefix = false
	pi.partialFindPrefix = false
	pi.headerVerified = false
	pi.prevPrefixCnt = 0
	pi.lastIndexPrefix = 0
	pi.currentPkgLen = 0
	pi.curPkgHeader = make([]byte, 0)
	pi.curPkg = make([]byte, 0)
}

func (pi *packetInspector) findPrefix(bb []byte) {
	//p := packetInspector{}
	if pi.completeFindPrefix {
		return
	}

	if pi.partialFindPrefix && pi.prevPrefixCnt > 0 {
		// There was a match with end of previous buffer
		// so we must find remaining of pattern in start of current buffer
		j := 0
		for i := pi.prevPrefixCnt; i < prefixLen; i++ {
			if j >= len(bb) {
				return
			}
			if packetPrefix[i] != bb[j] {
				break
			}
			pi.prevPrefixCnt++
			if pi.prevPrefixCnt == len(packetPrefix) {
				pi.completeFindPrefix = true
				pi.partialFindPrefix = false
				pi.prevPrefixCnt = 0
				pi.lastIndexPrefix = j
				break
			}
			j++
		}
	}
	// Patten does not exists partially in previous
	// buffer anf must find it in current buffer
	for i := 0; i < len(bb); i++ {
		if bb[i] == packetPrefix[0] {
			pi.partialFindPrefix = true
			pi.prevPrefixCnt = 1
			for j := 1; j < prefixLen; j++ {
				if j+i >= len(bb) {
					return
				}
				if bb[i+j] != packetPrefix[j] {
					pi.partialFindPrefix = false
					pi.prevPrefixCnt = 0
					break
				}
				pi.prevPrefixCnt++
				if pi.prevPrefixCnt == prefixLen {
					pi.completeFindPrefix = true
					pi.partialFindPrefix = false
					pi.prevPrefixCnt = 0
					pi.lastIndexPrefix = i + j
					return
				}
			}
		}
	}

}

func (pi *packetInspector) inspect(bb []byte, msgTypeLen map[byte]int) []rDataPacket {
	res := make([]rDataPacket, 0)
	if len(bb) == 0 {
		return res
	}
	dataStartIndex := 0
	if !pi.completeFindPrefix {
		pi.findPrefix(bb)
		if !pi.completeFindPrefix {
			return res
		}
		dataStartIndex = pi.lastIndexPrefix + 1
	}
	if dataStartIndex >= len(bb) {
		//aafter finding prefix we reach to the rnd of slice
		return res
	}
	if !pi.headerVerified {
		endOfHeader := 0
		if len(pi.curPkgHeader) < HeaderLen {
			// Incomplete header
			if (len(pi.curPkgHeader) + len(bb[dataStartIndex:len(bb)])) < HeaderLen {
				pi.curPkgHeader = append(pi.curPkgHeader, bb[dataStartIndex:len(bb)]...)
				return res
			}
			endOfHeader = dataStartIndex + (HeaderLen - len(pi.curPkgHeader))
			pi.curPkgHeader = append(pi.curPkgHeader, bb[dataStartIndex:endOfHeader]...)
		}
		msgTypeMaxLen, ok := msgTypeLen[pi.curPkgHeader[0]]
		currentPkgLen := int(binary.LittleEndian.Uint32(pi.curPkgHeader[1:]))

		if !ok || msgTypeMaxLen < currentPkgLen {
			// Message type or message len is not valid
			pi.resetVariables()
			res = append(res, pi.inspect(bb[dataStartIndex:], msgTypeLen)...)
			return res
		}
		pi.headerVerified = true
		pi.currentPkgLen = currentPkgLen
		dataStartIndex = endOfHeader
		if currentPkgLen <= len(bb[dataStartIndex:]) {
			pkt := rDataPacket{typ: pi.curPkgHeader[0], data: bb[dataStartIndex : dataStartIndex+currentPkgLen]}
			res = append(res, pkt)
			dataStartIndex += currentPkgLen
			if dataStartIndex > len(bb) {
				//aafter finding prefix we reach to the rnd of slice
				return res
			}
			pi.resetVariables()
			res = append(res, pi.inspect(bb[dataStartIndex:], msgTypeLen)...)
			return res
		}
		pi.curPkg = append(pi.curPkg, bb[dataStartIndex:]...)
		return res
	}
	// message prefix and header find previously
	// so just append to current package and search for next pckage
	if pi.currentPkgLen > len(pi.curPkg)+len(bb) {
		pi.curPkg = append(pi.curPkg, bb...)
		return res
	}
	remainLen := pi.currentPkgLen - len(pi.curPkg)
	pi.curPkg = append(pi.curPkg, bb[0:remainLen]...)
	pkt := rDataPacket{typ: pi.curPkgHeader[0], data: pi.curPkg}
	res = append(res, pkt)
	if remainLen >= len(bb) {
		//aafter finding prefix we reach to the rnd of slice
		return res
	}
	pi.resetVariables()
	res = append(res, pi.inspect(bb[remainLen:], msgTypeLen)...)
	return res
}
