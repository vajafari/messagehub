package message

import (
	"encoding/binary"
)

// IDRequestMsg represent request from client to get id from server
type IDRequestMsg struct {
}

// Type get type of id message
func (msg IDRequestMsg) Type() byte {
	return byte(IDMgsCode)
}

// Data get frame bytes of IDRequestMsg
func (msg IDRequestMsg) Data() ([]byte, error) {
	return nil, nil
}

// IDResponseMsg represent response of server to client and assign id to client
type IDResponseMsg struct {
	ID uint64
}

// Type get type of id message
func (msg IDResponseMsg) Type() byte {
	return byte(IDMgsCode)
}

// Data get frame bytes of IDResponseMsg
func (msg IDResponseMsg) Data() ([]byte, error) {
	res := make([]byte, 8)
	binary.LittleEndian.PutUint64(res, uint64(msg.ID))
	return res, nil
}

// DeserializeIDRes convert stream of bytes to IDResponseMsg
func DeserializeIDRes(bb []byte) (IDResponseMsg, error) {
	if len(bb) != 8 {
		return IDResponseMsg{}, ErrParsStream
	}

	return IDResponseMsg{
		ID: binary.LittleEndian.Uint64(bb),
	}, nil
}
