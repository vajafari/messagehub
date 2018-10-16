package message

import (
	"encoding/binary"
)

const (
	// ListMaxItems max length of data
	ListMaxItems int = 131072
)

// ListRequestMsg represent request from client to get list of connected clients
type ListRequestMsg struct {
}

// Type get type of list message
func (msg ListRequestMsg) Type() byte {
	return byte(ListMgsCode)
}

// Data get frame bytes of ListRequestMsg
func (msg ListRequestMsg) Data() ([]byte, error) {
	return nil, nil
}

// ListResponseMsg represent response of server to client return list of connected client
type ListResponseMsg struct {
	IDs []uint64
}

// Type get type of list message
func (msg ListResponseMsg) Type() byte {
	return byte(ListMgsCode)
}

// Data get frame bytes of ListRequestMsg
func (msg ListResponseMsg) Data() ([]byte, error) {
	if len(msg.IDs) > ListMaxItems {
		return nil, ErrInvalidData
	}
	return getUnit64Bytes(msg.IDs), nil
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
