package message

import "encoding/binary"

const (
	// MaxReciverCount max count of receivers per message
	MaxReciverCount int = 255
	// MaxBodySize max length of data
	MaxBodySize int = 1048576
)

// RelayRequestMsg represent request from client to relay a message to ither clients
type RelayRequestMsg struct {
	IDs  []uint64
	Data []byte
}

// Type get type of Relay message
func (msg *RelayRequestMsg) Type() byte {
	return byte(RelayMgsCode)
}

// Length get length of body for RelayRequestMsg object
func (msg *RelayRequestMsg) Length() uint32 {
	if len(msg.IDs) == 0 || len(msg.IDs) > MaxReciverCount {
		return 0
	}
	if len(msg.Data) == 0 || len(msg.Data) > MaxBodySize {
		return 0
	}

	return uint32(1 + (len(msg.IDs) * 8) + len(msg.Data)) // 1 bytes for count of recievers
}

// Serialize get frame bytes of ListRequestMsg
func (msg *RelayRequestMsg) Serialize() []byte {
	if len(msg.IDs) == 0 || len(msg.IDs) > MaxReciverCount {
		return nil
	}
	if len(msg.Data) == 0 || len(msg.Data) > MaxBodySize {
		return nil
	}

	hdr := makeHeaderBytes(msg.Type(), msg.Length())
	body := make([]byte, msg.Length())
	body[0] = byte(len(msg.IDs))
	copy(body[1:], getUnit64Bytes(msg.IDs))
	copy(body[1+(len(msg.IDs)*8):], msg.Data)
	return appendSlices(hdr, body)
}

// RelayResponseMsg represent message from server to clients
type RelayResponseMsg struct {
	SenderID uint64
	Data     []byte
}

// Type get type of Relay message
func (msg *RelayResponseMsg) Type() byte {
	return byte(RelayMgsCode)
}

// Length get length of body for RelayResponseMsg object
func (msg *RelayResponseMsg) Length() uint32 {
	if len(msg.Data) == 0 || len(msg.Data) > MaxBodySize {
		return 0
	}
	return uint32(len(msg.Data) + 8)
}

// Serialize get frame bytes of ListRequestMsg
func (msg *RelayResponseMsg) Serialize() []byte {

	if len(msg.Data) == 0 || len(msg.Data) > MaxBodySize {
		return nil
	}

	hdr := makeHeaderBytes(msg.Type(), msg.Length())
	body := make([]byte, msg.Length())
	bb := make([]byte, 8)
	binary.LittleEndian.PutUint64(bb, uint64(msg.SenderID))
	copy(body[0:], bb)
	copy(body[8:], msg.Data)
	return appendSlices(hdr, body)
}
