package message

import "testing"

func TestCheckEq(t *testing.T) {
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
		actual := checkEq(tt.first, tt.second)
		if checkEq(tt.first, tt.second) != tt.expected {
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
		if !checkEq(actual, tt.expected) {
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
		if !checkEq(actual, tt.expected) {
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
		if !checkEq(actual, tt.expected) {
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
		if !checkEq(actual, tt.expected) {
			t.Errorf("GetUnit64Bytes: expected %v, actual %v", tt.expected, actual)
		}
	}
}
