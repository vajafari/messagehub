package socket

//Socket define standard for communication channel
type Socket interface {
	SetCommChannels(chan<- WData, chan<- RData, chan<- ProbData)
	Send(frm Frame)
	Close() error
	ID() uint64
	SetID(uint64)
}
