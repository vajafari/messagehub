package message

import "errors"

// MsgType is type for Defining defferent message types
type MsgType byte

var (
	// For connection setup operations.
	ErrParsStream = errors.New("Stream cannot parse to message")
)

const (
	// IDMgsCode is code for id messages
	IDMgsCode MsgType = 1
	// ListMgsCode is code for id messages
	ListMgsCode MsgType = 2
	// RelayMgsCode is code for id messages
	RelayMgsCode MsgType = 3
)

type messager interface {
	Type() byte
	Length() uint32
	Serialize() []byte
}
