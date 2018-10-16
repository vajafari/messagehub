package message

import "encoding/binary"

const (
	// RelayMaxReciverCount max count of receivers per message
	RelayMaxReciverCount int = 255
	// RelayMaxBodySize max length of data
	RelayMaxBodySize int = 1048576
)

// RelayRequestMsg represent request from client to relay a message to ither clients
type RelayRequestMsg struct {
	IDs  []uint64
	Body []byte
}

// Type get type of Relay message
func (msg RelayRequestMsg) Type() byte {
	return byte(RelayMgsCode)
}

// Data get frame bytes of ListRequestMsg
func (msg RelayRequestMsg) Data() ([]byte, error) {
	if len(msg.IDs) == 0 || len(msg.IDs) > RelayMaxReciverCount {
		return nil, ErrInvalidData
	}
	if len(msg.Body) == 0 || len(msg.Body) > RelayMaxBodySize {
		return nil, ErrInvalidData
	}
	data := make([]byte, (len(msg.IDs)*8)+len(msg.Body)+1)
	data[0] = byte(len(msg.IDs))
	copy(data[1:], getUnit64Bytes(msg.IDs))
	copy(data[(len(msg.IDs)*8)+1:], msg.Body)
	return data, nil
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
		Body: bb[(bb[0]*8)+1:],
	}, nil
}

// RelayResponseMsg represent message from server to clients
type RelayResponseMsg struct {
	SenderID uint64
	Body     []byte
}

// Type get type of Relay message
func (msg RelayResponseMsg) Type() byte {
	return byte(RelayMgsCode)
}

// Data get frame bytes of ListRequestMsg
func (msg RelayResponseMsg) Data() ([]byte, error) {

	if len(msg.Body) == 0 || len(msg.Body) > RelayMaxBodySize {
		return nil, ErrInvalidData
	}

	data := make([]byte, 8+len(msg.Body))
	bb := make([]byte, 8)
	binary.LittleEndian.PutUint64(bb, uint64(msg.SenderID))
	copy(data[0:], bb)
	copy(data[8:], msg.Body)
	return data, nil
}

// DeserializeRelayRes convert stream of bytes to RelayResponseMsg
func DeserializeRelayRes(bb []byte) (RelayResponseMsg, error) {
	// 8 byte for sender id and at least one byte for data
	if len(bb) < 9 {
		return RelayResponseMsg{}, ErrParsStream
	}

	return RelayResponseMsg{
		SenderID: binary.LittleEndian.Uint64(bb[0:8]),
		Body:     bb[8:],
	}, nil
}
