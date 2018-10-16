package socket

// HeaderLen is Length for header of frames
const HeaderLen int = 5

// Packet define standrd for message type
type Packet interface {
	Type() byte
	Data() ([]byte, error)
}
