package socket

// WData hold inforamtion wrote frame result
type WData struct {
	Frm      Frame
	SourceID uint64
}

// RData Hold information about input data
type RData struct {
	MsgType  byte
	Data     []byte
	SourceID uint64
}

// ProbData hold inforamtion problem on socket
type ProbData struct {
	Frm      Frame
	SourceID uint64
	Err      error
}
