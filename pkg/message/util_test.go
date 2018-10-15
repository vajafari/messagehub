package message

import "testing"

func TestCheckEqByte(t *testing.T) {
	var tests = []struct {
		first    []byte
		second   []byte
		expected bool // expected result
	}{
		{nil, nil, true},
		{nil, []byte{}, true},
		{[]byte{}, nil, true},
		{nil, []byte{1}, false},
		{[]byte{1}, nil, false},
		{[]byte{}, []byte{1}, false},
		{[]byte{1}, []byte{}, false},
		{[]byte{1}, []byte{1, 2}, false},
		{[]byte{1, 2}, []byte{1}, false},
		{[]byte{1, 2, 3}, []byte{1, 2, 4}, false},
		{[]byte{1, 2, 3}, []byte{1, 2, 3}, true},
	}
	for _, tt := range tests {
		actual := checkEqByte(tt.first, tt.second)
		if checkEqByte(tt.first, tt.second) != tt.expected {
			t.Errorf("checkEq: first %v second %v expected %t, actual %t", tt.first, tt.second, tt.expected, actual)
		}
	}
}

func TestMakeHeaderBytes(t *testing.T) {
	var tests = []struct {
		messageType byte
		length      uint32
		expected    []byte // expected result
	}{
		{0, 0, []byte{0, 0, 0, 0, 0}},
		{1, 0, []byte{1, 0, 0, 0, 0}},
		{10, 256, []byte{10, 0, 1, 0, 0}},
		{186, 986458765, []byte{186, 141, 42, 204, 58}},
	}
	for _, tt := range tests {
		actual := makeHeaderBytes(tt.messageType, tt.length)
		if !checkEqByte(actual, tt.expected) {
			t.Errorf("makeHeaderBytes: expected %v, actual %v", tt.expected, actual)
		}
	}
}

func TestAppendSlices(t *testing.T) {
	var tests = []struct {
		first    []byte // input
		second   []byte // input
		expected []byte // expected result
	}{
		{nil, nil, nil},
		{nil, []byte{0, 1, 2, 3}, []byte{0, 1, 2, 3}},
		{[]byte{0, 1, 2, 3}, nil, []byte{0, 1, 2, 3}},
		{[]byte{}, []byte{0, 1, 2, 3}, []byte{0, 1, 2, 3}},
		{[]byte{0, 1, 2, 3}, []byte{}, []byte{0, 1, 2, 3}},
		{[]byte{0, 1, 2, 3}, []byte{0, 1, 2, 3}, []byte{0, 1, 2, 3, 0, 1, 2, 3}},
		{[]byte{0, 1}, []byte{0, 1, 2, 3}, []byte{0, 1, 0, 1, 2, 3}},
	}
	for _, tt := range tests {
		actual := appendSlices(tt.first, tt.second)
		if !checkEqByte(actual, tt.expected) {
			t.Errorf("appendSlices: expected %v, actual %v", tt.expected, actual)
		}
	}
}

func TestGetUnit32Bytes(t *testing.T) {
	var tests = []struct {
		nums     []uint32 // input
		expected []byte   // expected result
	}{
		{nil, nil},
		{[]uint32{}, nil},
		{[]uint32{256}, []byte{0, 1, 0, 0}},
		{[]uint32{3, 27, 492, 4587, 87345, 159743, 1468743, 22446078, 374557900, 3459980326}, []byte{3, 0, 0, 0, 27, 0, 0, 0, 236, 1, 0, 0, 235, 17, 0, 0, 49, 85, 1, 0, 255, 111, 2, 0, 71, 105, 22, 0, 254, 127, 86, 1, 204, 76, 83, 22, 38, 28, 59, 206}},
	}
	for _, tt := range tests {
		actual := getUnit32Bytes(tt.nums)
		if !checkEqByte(actual, tt.expected) {
			t.Errorf("getUnit32Bytes: expected %v, actual %v", tt.expected, actual)
		}
	}
}

func TestGetUnit64Bytes(t *testing.T) {
	var tests = []struct {
		nums     []uint64 // input
		expected []byte   // expected result
	}{
		{nil, nil},
		{[]uint64{}, nil},
		{[]uint64{256}, []byte{0, 1, 0, 0, 0, 0, 0, 0}},
		{[]uint64{3, 27, 492, 4587, 87345, 159743, 1468743, 22446078, 374557900, 3459980326, 348529584035, 3849560104835, 39475690128456},
			[]byte{3, 0, 0, 0, 0, 0, 0, 0, 27, 0, 0, 0, 0, 0, 0, 0, 236, 1, 0, 0, 0, 0, 0, 0, 235, 17, 0, 0, 0, 0, 0, 0, 49, 85, 1, 0, 0, 0, 0, 0, 255, 111, 2, 0, 0, 0, 0, 0, 71, 105, 22, 0, 0, 0, 0, 0, 254, 127, 86, 1, 0, 0, 0, 0, 204, 76, 83, 22, 0, 0, 0, 0, 38, 28, 59, 206, 0, 0, 0, 0, 163, 103, 251,
				37, 81, 0, 0, 0, 131, 159, 169, 75, 128, 3, 0, 0, 72, 196, 124, 38, 231, 35, 0, 0}},
	}
	for _, tt := range tests {
		actual := getUnit64Bytes(tt.nums)
		if !checkEqByte(actual, tt.expected) {
			t.Errorf("GetUnit64Bytes: expected %v, actual %v", tt.expected, actual)
		}
	}
}

func TestChkIDResponseMsgEq(t *testing.T) {
	var tests = []struct {
		a        IDResponseMsg // input
		b        IDResponseMsg // input
		expected bool          // expected result
	}{
		{IDResponseMsg{}, IDResponseMsg{}, true},
		{IDResponseMsg{ID: 1}, IDResponseMsg{ID: 1}, true},
		{IDResponseMsg{ID: 1}, IDResponseMsg{ID: 2}, false},
	}
	for _, tt := range tests {
		actual := ChkIDResponseMsgEq(tt.a, tt.b)
		if actual != tt.expected {
			t.Errorf("ChkIDResponseMsgEq: expected %t, actual %t", tt.expected, actual)
		}
	}
}

func TestChkListResponseMsgEq(t *testing.T) {
	var tests = []struct {
		a        ListResponseMsg // input
		b        ListResponseMsg // input
		expected bool            // expected result
	}{
		{ListResponseMsg{}, ListResponseMsg{}, true},
		{ListResponseMsg{IDs: []uint64{}}, ListResponseMsg{IDs: []uint64{}}, true},
		{ListResponseMsg{}, ListResponseMsg{IDs: []uint64{}}, true},
		{ListResponseMsg{IDs: []uint64{}}, ListResponseMsg{}, true},
		{ListResponseMsg{}, ListResponseMsg{IDs: []uint64{1}}, false},
		{ListResponseMsg{IDs: []uint64{1}}, ListResponseMsg{}, false},
		{ListResponseMsg{IDs: []uint64{1, 2, 3, 4, 5}}, ListResponseMsg{IDs: []uint64{1, 2, 3, 4, 6}}, false},
		{ListResponseMsg{IDs: []uint64{1, 2, 3, 4, 5}}, ListResponseMsg{IDs: []uint64{1, 2, 3}}, false},
		{ListResponseMsg{IDs: []uint64{1, 2, 3}}, ListResponseMsg{IDs: []uint64{1, 2, 3, 4, 5}}, false},
		{ListResponseMsg{IDs: []uint64{1, 2, 3, 4, 5}}, ListResponseMsg{IDs: []uint64{1, 2, 3, 4, 5}}, true},
	}
	for _, tt := range tests {
		actual := ChkListResponseMsgEq(tt.a, tt.b)
		if actual != tt.expected {
			t.Errorf("ChkListResponseMsgEq: expected %t, actual %t", tt.expected, actual)
		}
	}
}

func TestChkRelayRequestMsgEq(t *testing.T) {
	var tests = []struct {
		a        RelayRequestMsg // input
		b        RelayRequestMsg // input
		expected bool            // expected result
	}{
		{RelayRequestMsg{IDs: []uint64{}, Data: []byte{}}, RelayRequestMsg{IDs: []uint64{}, Data: []byte{}}, true},
		{RelayRequestMsg{Data: []byte{}}, RelayRequestMsg{IDs: []uint64{}, Data: []byte{}}, true},
		{RelayRequestMsg{IDs: []uint64{}}, RelayRequestMsg{IDs: []uint64{}, Data: []byte{}}, true},
		{RelayRequestMsg{IDs: []uint64{}, Data: []byte{}}, RelayRequestMsg{Data: []byte{}}, true},
		{RelayRequestMsg{IDs: []uint64{}, Data: []byte{}}, RelayRequestMsg{IDs: []uint64{}}, true},
		{RelayRequestMsg{}, RelayRequestMsg{IDs: []uint64{}, Data: []byte{}}, true},
		{RelayRequestMsg{IDs: []uint64{}}, RelayRequestMsg{Data: []byte{}}, true},
		{RelayRequestMsg{IDs: []uint64{}, Data: []byte{}}, RelayRequestMsg{}, true},
		{RelayRequestMsg{}, RelayRequestMsg{Data: []byte{}}, true},
		{RelayRequestMsg{IDs: []uint64{}}, RelayRequestMsg{}, true},
		{RelayRequestMsg{}, RelayRequestMsg{}, true},
		{RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, true},
		{RelayRequestMsg{Data: []byte{4, 5, 6}}, RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, false},
		{RelayRequestMsg{IDs: []uint64{1, 2, 3}}, RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, false},
		{RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, RelayRequestMsg{Data: []byte{4, 5, 6}}, false},
		{RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, RelayRequestMsg{IDs: []uint64{1, 2, 3}}, false},
		{RelayRequestMsg{}, RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, false},
		{RelayRequestMsg{IDs: []uint64{1, 2, 3}}, RelayRequestMsg{Data: []byte{4, 5, 6}}, false},
		{RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, RelayRequestMsg{}, false},
		{RelayRequestMsg{}, RelayRequestMsg{Data: []byte{4, 5, 6}}, false},
		{RelayRequestMsg{IDs: []uint64{1, 2, 3}}, RelayRequestMsg{}, false},
		{RelayRequestMsg{IDs: []uint64{1, 2, 4}, Data: []byte{4, 5, 6}}, RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, false},
		{RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 7}}, RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, false},
		{RelayRequestMsg{IDs: []uint64{1, 2, 3, 4}, Data: []byte{4, 5, 6}}, RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, false},
		{RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6}}, RelayRequestMsg{IDs: []uint64{1, 2, 3}, Data: []byte{4, 5, 6, 7}}, false},
		{RelayRequestMsg{IDs: []uint64{3, 5, 7}, Data: []byte{1, 3, 5}}, RelayRequestMsg{IDs: []uint64{2, 4, 6}, Data: []byte{6, 8, 9}}, false},
	}
	for _, tt := range tests {
		actual := ChkRelayRequestMsgEq(tt.a, tt.b)
		if actual != tt.expected {
			t.Errorf("ChkRelayRequestMsgEq: expected %t, actual %t", tt.expected, actual)
		}
	}
}

func TestChkRelayResponseMsgEq(t *testing.T) {
	var tests = []struct {
		a        RelayResponseMsg // input
		b        RelayResponseMsg // input
		expected bool             // expected result
	}{
		{RelayResponseMsg{}, RelayResponseMsg{}, true},
		{RelayResponseMsg{Data: []byte{}}, RelayResponseMsg{Data: []byte{}}, true},
		{RelayResponseMsg{}, RelayResponseMsg{Data: []byte{}}, true},
		{RelayResponseMsg{Data: []byte{}}, RelayResponseMsg{}, true},
		{RelayResponseMsg{}, RelayResponseMsg{Data: []byte{1}}, false},
		{RelayResponseMsg{Data: []byte{1}}, RelayResponseMsg{}, false},
		{RelayResponseMsg{Data: []byte{1, 2, 3, 4, 5}}, RelayResponseMsg{Data: []byte{1, 2, 3, 4, 6}}, false},
		{RelayResponseMsg{Data: []byte{1, 2, 3, 4, 5}}, RelayResponseMsg{Data: []byte{1, 2, 3}}, false},
		{RelayResponseMsg{Data: []byte{1, 2, 3}}, RelayResponseMsg{Data: []byte{1, 2, 3, 4, 5}}, false},
		{RelayResponseMsg{Data: []byte{1, 2, 3, 4, 5}}, RelayResponseMsg{Data: []byte{1, 2, 3, 4, 5}}, true},
	}
	for _, tt := range tests {
		actual := ChkRelayResponseMsgEq(tt.a, tt.b)
		if actual != tt.expected {
			t.Errorf("ChkRelayResponseMsgEq: expected %t, actual %t", tt.expected, actual)
		}
	}
}
