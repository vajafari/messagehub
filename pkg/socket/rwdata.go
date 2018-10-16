package socket

// WData hold inforamtion wrote frame result
type WData struct {
	Pkt      Packet
	SourceID uint64
}

// RData Hold information about input data
type RData struct {
	Pkt      Packet
	SourceID uint64
}

// ProbData hold inforamtion problem on socket
type ProbData struct {
	Pkt      Packet
	SourceID uint64
	Err      error
}

type rDataPacket struct {
	typ  byte
	data []byte
}

func (pkt rDataPacket) Type() byte {
	return pkt.typ
}
func (pkt rDataPacket) Data() ([]byte, error) {
	return pkt.data, nil
}
