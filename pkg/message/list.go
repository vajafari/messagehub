package message

import "encoding/binary"

// ListRequestMsg represent request from client to get list of connected clients
type ListRequestMsg struct {
}

// Type get type of list message
func (msg ListRequestMsg) Type() byte {
	return byte(ListMgsCode)
}

// Length get length of body for ListRequestMsg object
func (msg ListRequestMsg) Length() uint32 {
	return 0
}

// Serialize get frame bytes of ListRequestMsg
func (msg ListRequestMsg) Serialize() []byte {
	return appendSlices(makeHeaderBytes(msg.Type(), msg.Length()), nil)
}

// ListResponseMsg represent response of server to client return list of connected client
type ListResponseMsg struct {
	IDs []uint64
}

// Type get type of list message
func (msg ListResponseMsg) Type() byte {
	return byte(ListMgsCode)
}

// Length get length of body for ListResponseMsg object
func (msg ListResponseMsg) Length() uint32 {
	if len(msg.IDs) > 0 {
		return uint32(len(msg.IDs) * 8)
	}
	return 0
}

// Serialize get frame bytes of ListRequestMsg
func (msg ListResponseMsg) Serialize() []byte {
	return appendSlices(makeHeaderBytes(msg.Type(), msg.Length()), getUnit64Bytes(msg.IDs))
}

// DeserializeListRes convert stream of bytes to ListResponseMsg
func DeserializeListRes(bb []byte) (ListResponseMsg, error) {
	if len(bb)%8 != 0 {
		return ListResponseMsg{}, ErrParsStream
	}
	if len(bb) == 0 {
		return ListResponseMsg{}, nil
	}
	cnt := len(bb) / 8
	uu := make([]uint64, cnt)
	for i := 0; i < cnt; i++ {
		uu[i] = binary.LittleEndian.Uint64(bb[i*8 : ((i + 1) * 8)])
	}

	return ListResponseMsg{
		IDs: uu,
	}, nil
}
