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
func (msg RelayRequestMsg) Type() byte {
	return byte(RelayMgsCode)
}

// Length get length of body for RelayRequestMsg object
func (msg RelayRequestMsg) Length() uint32 {
	if len(msg.IDs) == 0 || len(msg.IDs) > MaxReciverCount {
		return 0
	}
	if len(msg.Data) == 0 || len(msg.Data) > MaxBodySize {
		return 0
	}

	return uint32(1 + (len(msg.IDs) * 8) + len(msg.Data)) // 1 bytes for count of recievers
}

// Serialize get frame bytes of ListRequestMsg
func (msg RelayRequestMsg) Serialize() []byte {
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

// DeserializeRelayReq convert stream of bytes to RelayRequestMsg
func DeserializeRelayReq(bb []byte) (RelayRequestMsg, error) {
	// 1 byte for reciever list len, 8 byte for at leat one reciever and at least one byte for data
	if len(bb) < 10 {
		return RelayRequestMsg{}, ErrParsStream
	}

	// 1 byte for reciever list len, byte[0] * 8 byte for recievers and at least one byte for data
	if len(bb) < int(bb[0]*8)+2 {
		return RelayRequestMsg{}, ErrParsStream
	}

	if bb[0] == 0 {
		return RelayRequestMsg{}, ErrParsStream
	}

	uu := make([]uint64, bb[0])
	cnt := int(bb[0])
	for i := 0; i < cnt; i++ {
		uu[i] = binary.LittleEndian.Uint64(bb[(i*8)+1 : ((i+1)*8)+1])
	}

	return RelayRequestMsg{
		IDs:  uu,
		Data: bb[(bb[0]*8)+1:],
	}, nil
}

// RelayResponseMsg represent message from server to clients
type RelayResponseMsg struct {
	SenderID uint64
	Data     []byte
}

// Type get type of Relay message
func (msg RelayResponseMsg) Type() byte {
	return byte(RelayMgsCode)
}

// Length get length of body for RelayResponseMsg object
func (msg RelayResponseMsg) Length() uint32 {
	if len(msg.Data) == 0 || len(msg.Data) > MaxBodySize {
		return 0
	}
	return uint32(len(msg.Data) + 8)
}

// Serialize get frame bytes of ListRequestMsg
func (msg RelayResponseMsg) Serialize() []byte {

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

// DeserializeRelayRes convert stream of bytes to RelayResponseMsg
func DeserializeRelayRes(bb []byte) (RelayResponseMsg, error) {
	// 8 byte for sender id and at least one byte for data
	if len(bb) < 9 {
		return RelayResponseMsg{}, ErrParsStream
	}

	return RelayResponseMsg{
		SenderID: binary.LittleEndian.Uint64(bb[0:8]),
		Data:     bb[8:],
	}, nil
}
