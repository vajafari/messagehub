package socket

//Socket define standard for communication channel
type Socket interface {
	Start(chan<- WData, chan<- RData, chan<- ProbData, map[byte]int)
	Close() error
	ID() uint64
	SetID(uint64)
	Send(frm Packet)
}
