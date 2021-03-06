package socket

import (
	"bufio"
	"encoding/binary"
	"fmt"
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

// TCPSocket holds information a connection between client and server
// This class designed to be reusable across projects
type TCPSocket struct {
	conn      *net.TCPConn    // TCP connection.
	id        uint64          // Assigned ID to current TCPSocket
	sendQueue chan Packet     // Outgoing packets queue. We use a buffered channel of packets as thread-safe FIFO queue
	closeGoes chan bool       // This channel used to stop all go routines of TCPSocekt
	readChan  chan<- RData    // if successful read happen signal send through this channel
	writeChan chan<- WData    // if successful write happen signal send through this channel
	probChan  chan<- ProbData // if error occur signal to send through this channel
	// This is a map that specifies how much data is valid for each type of message
	// We use this to prevent the client from sending irrational data.
	// Each packet type (first byte of packet) has max length
	msgTypeLen   map[byte]int
	readBufSize  int
	writeBufSize int
}

//NewTCPSocket create TCP Socket object to hold client collection info
func NewTCPSocket(conn *net.TCPConn, id uint64, sendQueueSize int, readBufSize int, writeBufSize int) *TCPSocket {
	s := TCPSocket{
		conn:         conn,
		sendQueue:    make(chan Packet, sendQueueSize),
		closeGoes:    make(chan bool, 1),
		readBufSize:  readBufSize,
		writeBufSize: writeBufSize,
		id:           id,
	}
	return &s
}

//Start set channels to communicate with the socket manager
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
}

//Send Add packet to send queue
func (s *TCPSocket) Send(pkt Packet) {
	s.sendQueue <- pkt
}

// Close tcpSocket and release all the resources
func (s *TCPSocket) Close() error {
	err := s.conn.Close()
	if err == nil {
		s.closeGoes <- true
		s.closeGoes <- true
		close(s.sendQueue)
		close(s.closeGoes)
	} else {
		fmt.Println(err)
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
}

func (s *TCPSocket) writer() {
	bufio.NewWriterSize(s.conn, s.writeBufSize)
	buf := bufio.NewWriter(s.conn)
	for {
		//When we close 'send' channel, It still received buffered packets so
		//writer go routine continue to work. we use the following select to make sure that
		//this go routine closes immediately when we call TCPSocket close method
		select {
		case <-s.closeGoes:
			return
		case pkt := <-s.sendQueue:
			if pkt == nil {
				continue
			}
			// Prepare data for sending on wire!!!
			bb, err := pkt.Data()
			if err != nil {
				continue
			}
			pktBytes := make([]byte, prefixLen+HeaderLen+len(bb))
			copy(pktBytes, packetPrefix)     //Prefix
			pktBytes[prefixLen] = pkt.Type() //Type
			lenBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(lenBytes, uint32(len(bb)))
			copy(pktBytes[prefixLen+1:], lenBytes) //Lent
			copy(pktBytes[prefixLen+HeaderLen:], bb)
			// Data composed
			_, err = s.writeWithRetry(buf, pktBytes, writeTimeout)
			if err != nil {
				fmt.Printf("TCPSocket, Error on send data--- Socket%d   %s\n", s.id, err.Error())
				res := ProbData{
					Pkt:      pkt,
					SourceID: s.id,
					Err:      err,
				}
				s.probChan <- res
				return
			}
			res := WData{
				Pkt:      pkt,
				SourceID: s.id,
			}
			//fmt.Printf("TcpSocket, Data Sent-- Socket %d- Len %d\n", s.id, nn)
			s.writeChan <- res

		}
	}
}

func (s *TCPSocket) writeWithRetry(buf *bufio.Writer, bb []byte, timeout time.Duration) (int, error) {
	s.conn.SetWriteDeadline(time.Now().Add(timeout))
	nn, err := buf.Write(bb)
	if err != nil {
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
	bb := make([]byte, s.readBufSize)
	buf := bufio.NewReaderSize(s.conn, s.readBufSize)
	pi := packetInspector{}
	pi.resetVariables()
	for {
		select {
		case <-s.closeGoes:
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
			return
		}
		if n == 0 {
			continue
		}
		packets := pi.inspect(bb[0:n], s.msgTypeLen)
		//fmt.Printf("LEN PACKETS:%d---CFP:%t--PFP:%t--HV:%t--PPC:%d--LIP:%d--CPL:%d--CPH:%d--CP:%d\n",
		//	len(packets), pi.completeFindPrefix, pi.partialFindPrefix, pi.headerVerified,
		//	pi.prevPrefixCnt, pi.lastIndexPrefix, pi.currentPkgLen, len(pi.curPkgHeader), len(pi.curPkg))
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
		// There was a match with the end of the previous buffer
		// so we must find remaining of pattern in the start of the current buffer
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
	// Patten does not exist partially in previous
	// buffer and must find it in the current buffer
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
		//after finding prefix we reach to the end of slice
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
				//After finding prefix we reach to the end of slice
				return res
			}
			pi.resetVariables()
			res = append(res, pi.inspect(bb[dataStartIndex:], msgTypeLen)...)
			return res
		}
		pi.curPkg = append(pi.curPkg, bb[dataStartIndex:]...)
		return res
	}
	// Message prefix and header find previously
	// so just append to current package and search for next package
	if pi.currentPkgLen > len(pi.curPkg)+len(bb) {
		pi.curPkg = append(pi.curPkg, bb...)
		return res
	}
	remainLen := pi.currentPkgLen - len(pi.curPkg)
	pi.curPkg = append(pi.curPkg, bb[0:remainLen]...)
	pkt := rDataPacket{typ: pi.curPkgHeader[0], data: pi.curPkg}
	res = append(res, pkt)
	// if remainLen >= len(bb) {
	// 	//aafter finding prefix we reach to the rnd of slice
	// 	return res
	// }
	pi.resetVariables()
	res = append(res, pi.inspect(bb[remainLen:], msgTypeLen)...)
	return res
}
