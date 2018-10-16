package message

import "encoding/binary"

// checkEqByte In this context nil sclice considered as empty slice, so they are equal
func checkEqByte(a, b []byte) bool {

	if len(a) != len(b) {
		return false
	}
	if len(a) > 0 {
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
	}
	return true
}
func checkEqUint64(a, b []uint64) bool {

	if len(a) != len(b) {
		return false
	}
	if len(a) > 0 {
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
	}
	return true
}

func appendSlices(s1 []byte, s2 []byte) []byte {
	bb := make([]byte, len(s1)+len(s2))
	if len(s1) > 0 {
		copy(bb[0:], s1)
	}
	if len(s2) > 0 {
		copy(bb[len(s1):], s2)
	}
	return bb
}

func getUnit32Bytes(input []uint32) []byte {
	if len(input) == 0 {
		return nil
	}
	bb := make([]byte, 4)

	res := make([]byte, len(input)*4)
	for i, n := range input {
		binary.LittleEndian.PutUint32(bb, n)
		copy(res[i*4:], bb)
	}
	return res
}

func getUnit64Bytes(input []uint64) []byte {
	if len(input) == 0 {
		return nil
	}
	bb := make([]byte, 8)

	res := make([]byte, len(input)*8)
	for i, n := range input {
		binary.LittleEndian.PutUint64(bb, n)
		copy(res[i*8:], bb)
	}
	return res
}

// ChkIDResponseMsgEq check equeality of IDResponseMsg message
func ChkIDResponseMsgEq(a, b IDResponseMsg) bool {
	return a.ID == b.ID
}

// ChkListResponseMsgEq check equeality of ListResponseMsg message
func ChkListResponseMsgEq(a, b ListResponseMsg) bool {
	return checkEqUint64(a.IDs, b.IDs)
}

// ChkRelayRequestMsgEq check equeality of RelayRequestMsg message
func ChkRelayRequestMsgEq(a, b RelayRequestMsg) bool {
	return checkEqUint64(a.IDs, b.IDs) && checkEqByte(a.Body, b.Body)
}

// ChkRelayResponseMsgEq check equeality of RelayResponseMsg message
func ChkRelayResponseMsgEq(a, b RelayResponseMsg) bool {
	return checkEqByte(a.Body, b.Body) && a.SenderID == b.SenderID
}
