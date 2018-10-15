package message

import "errors"

// MsgType is type for Defining defferent message types
type MsgType byte

var (
	// ErrParsStream reporesent error on deserilize stream
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
