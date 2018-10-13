package message

import (
	"encoding/binary"
)

// IDRequestMsg represent request from client to get id from server
type IDRequestMsg struct {
}

// Type get type of id message
func (msg *IDRequestMsg) Type() byte {
	return byte(IDMgsCode)
}

// Length get length of body for IDRequestMsg object
func (msg *IDRequestMsg) Length() uint32 {
	return 0
}

// Serialize get frame bytes of IDRequestMsg
func (msg *IDRequestMsg) Serialize() []byte {
	return appendSlices(makeHeaderBytes(msg.Type(), msg.Length()), nil)
}

// IDResponseMsg represent response of server to client and assign id to client
type IDResponseMsg struct {
	ID uint64
}

// Type get type of id message
func (msg *IDResponseMsg) Type() byte {
	return byte(IDMgsCode)
}

// Length get length of body for IDResponseMsg object
func (msg *IDResponseMsg) Length() uint32 {
	return 8
}

// Serialize get frame bytes of IDResponseMsg
func (msg *IDResponseMsg) Serialize() []byte {
	res := make([]byte, 13)
	res[0] = msg.Type()
	res[1] = 8
	bb := make([]byte, 8)
	binary.LittleEndian.PutUint64(bb, uint64(msg.ID))
	copy(res[5:], bb)
	return res
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
