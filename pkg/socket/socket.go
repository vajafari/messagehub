package socket

//Socket define standard for communication channel
type Socket interface {
	Start(chan<- WData, chan<- RData, chan<- ProbData, map[byte]uint32)
	Close() error
	ID() uint64
	SetID(uint64)
	Send(frm Frame)
}
