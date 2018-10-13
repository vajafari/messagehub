package socket

// HeaderLen is Length for header of frames
const HeaderLen int = 5

// Frame define standrd for message type
type Frame interface {
	Type() byte
	Length() uint32
	Serialize() []byte
}
