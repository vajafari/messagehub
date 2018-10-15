package message

import (
	"testing"
)

type messager interface {
	Type() byte
	Length() uint32
	Serialize() []byte
}

func TestType(t *testing.T) {

	var tests = []struct {
		msg      messager // input
		typeName string
		expected MsgType // expected result
	}{
		{&IDRequestMsg{}, "IDRequestMsg", IDMgsCode},
		{&IDResponseMsg{}, "IDResponseMsg", IDMgsCode},
		{&ListRequestMsg{}, "ListRequestMsg", ListMgsCode},
		{&ListResponseMsg{}, "ListResponseMsg", ListMgsCode},
		{&RelayRequestMsg{}, "RelayRequestMsg", RelayMgsCode},
		{&RelayResponseMsg{}, "RelayResponseMsg", RelayMgsCode},
	}
	for _, tt := range tests {
		actual := tt.msg.Type()
		if actual != byte(tt.expected) {
			t.Errorf("%s.Type: expected %d, actual %d", tt.typeName, tt.expected, actual)
		}
	}
}

func TestLength(t *testing.T) {

	// bigConnListResp represnet message
	bigListResp := ListResponseMsg{}
	bigListResp.IDs = make([]uint64, 100)
	for i := 0; i < 100; i++ {
		bigListResp.IDs[i] = 123
	}

	tooMuchIDRelay := RelayRequestMsg{}
	tooMuchIDRelay.IDs = make([]uint64, 300)
	for i := 0; i < 300; i++ {
		tooMuchIDRelay.IDs[i] = 123
	}

	tooMuchDataRelay := RelayRequestMsg{}
	tooMuchDataRelay.Data = make([]byte, MaxBodySize*2)
	for i := 0; i < MaxBodySize*2; i++ {
		tooMuchDataRelay.Data[i] = 123
	}

	tooMuchDataRelayResp := RelayResponseMsg{}
	tooMuchDataRelayResp.Data = make([]byte, MaxBodySize*2)
	for i := 0; i < MaxBodySize*2; i++ {
		tooMuchDataRelayResp.Data[i] = 123
	}

	var tests = []struct {
		msg      messager // input
		typeName string
		expected uint32 // expected result
	}{
		{&IDRequestMsg{}, "IDRequestMsg", 0},
		{&IDResponseMsg{}, "IDResponseMsg", 8},

		{&ListRequestMsg{}, "ListRequestMsg", 0},

		{&ListResponseMsg{}, "ListResponseMsg", 0},
		{&ListResponseMsg{IDs: []uint64{}}, "ListResponseMsg", 0},
		{&ListResponseMsg{IDs: []uint64{1}}, "ListResponseMsg", 8},
		{&ListResponseMsg{IDs: []uint64{1, 2, 3}}, "ListResponseMsg", 24},
		{&bigListResp, "ListResponseMsg", 800},

		{&RelayRequestMsg{}, "RelayRequestMsg", 0},
		{&RelayRequestMsg{Data: []byte{}}, "RelayRequestMsg", 0},
		{&RelayRequestMsg{IDs: []uint64{}}, "RelayRequestMsg", 0},
		{&RelayRequestMsg{Data: []byte{}, IDs: []uint64{}}, "RelayRequestMsg", 0},
		{&RelayRequestMsg{Data: []byte{123, 124, 125}}, "RelayRequestMsg", 0},      // only data with no reciever
		{&RelayRequestMsg{IDs: []uint64{0, 1, 2, 3, 4, 54}}, "RelayRequestMsg", 0}, // only reciever with no data
		{&tooMuchIDRelay, "RelayRequestMsg", 0},
		{&tooMuchDataRelay, "RelayRequestMsg", 0},
		{&RelayRequestMsg{IDs: []uint64{1, 2}, Data: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}}, "RelayRequestMsg", 27},

		{&RelayResponseMsg{}, "RelayResponseMsg", 0},
		{&RelayResponseMsg{Data: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}}, "RelayRequestMsg", 18},
		{&tooMuchDataRelayResp, "RelayResponseMsg", 0},
	}

	for _, tt := range tests {
		actual := tt.msg.Length()
		if actual != uint32(tt.expected) {
			t.Errorf("%s.Type: expected %d, actual %d", tt.typeName, tt.expected, actual)
		}
	}
}

func TestSerialize(t *testing.T) {

	// bigConnListResp represnet message
	bigListResp := ListResponseMsg{}
	bigListResp.IDs = make([]uint64, 100)
	for i := 0; i < 100; i++ {
		bigListResp.IDs[i] = 123
	}

	tooMuchIDRelay := RelayRequestMsg{}
	tooMuchIDRelay.IDs = make([]uint64, 300)
	for i := 0; i < 300; i++ {
		tooMuchIDRelay.IDs[i] = 123
	}

	tooMuchDataRelay := RelayRequestMsg{}
	tooMuchDataRelay.Data = make([]byte, MaxBodySize*2)
	for i := 0; i < MaxBodySize*2; i++ {
		tooMuchDataRelay.Data[i] = 123
	}

	tooMuchDataRelayResp := RelayResponseMsg{}
	tooMuchDataRelayResp.Data = make([]byte, MaxBodySize*2)
	for i := 0; i < MaxBodySize*2; i++ {
		tooMuchDataRelayResp.Data[i] = 123
	}

	var tests = []struct {
		msg      messager // input
		typeName string
		expected []byte // expected result
	}{
		{&IDRequestMsg{}, "IDRequestMsg", []byte{1, 0, 0, 0, 0}},
		{&IDResponseMsg{ID: 256}, "IDResponseMsg", []byte{1, 8, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0}},

		{&ListRequestMsg{}, "ListRequestMsg", []byte{2, 0, 0, 0, 0}},

		{&ListResponseMsg{}, "ListResponseMsg", []byte{2, 0, 0, 0, 0}},
		{&ListResponseMsg{IDs: []uint64{}}, "ListResponseMsg", []byte{2, 0, 0, 0, 0}},
		{&ListResponseMsg{IDs: []uint64{1}}, "ListResponseMsg", []byte{2, 8, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}},
		{&ListResponseMsg{IDs: []uint64{1, 2, 3}}, "ListResponseMsg", []byte{2, 24, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0}},

		{&RelayRequestMsg{}, "RelayRequestMsg", nil},
		{&RelayRequestMsg{Data: []byte{}}, "RelayRequestMsg", nil},
		{&RelayRequestMsg{IDs: []uint64{}}, "RelayRequestMsg", nil},
		{&RelayRequestMsg{Data: []byte{}, IDs: []uint64{}}, "RelayRequestMsg", nil},
		{&RelayRequestMsg{Data: []byte{123, 124, 125}}, "RelayRequestMsg", nil},      // only data with no reciever
		{&RelayRequestMsg{IDs: []uint64{0, 1, 2, 3, 4, 54}}, "RelayRequestMsg", nil}, // only reciever with no data
		{&tooMuchIDRelay, "RelayRequestMsg", nil},
		{&tooMuchDataRelay, "RelayRequestMsg", nil},
		{&RelayRequestMsg{IDs: []uint64{1, 2}, Data: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}}, "RelayRequestMsg", []byte{3, 27, 0, 0, 0, 2, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},

		{&RelayResponseMsg{}, "RelayResponseMsg", nil},
		{&RelayResponseMsg{SenderID: 1, Data: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}}, "RelayRequestMsg", []byte{3, 18, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{&tooMuchDataRelayResp, "RelayResponseMsg", nil},
	}

	for _, tt := range tests {
		actual := tt.msg.Serialize()
		if !checkEqByte(actual, tt.expected) {
			t.Errorf("%s.Type: expected %v, actual %v", tt.typeName, tt.expected, actual)
		}
	}
}

func TestDeserializeIDRes(t *testing.T) {
	var tests = []struct {
		stream []byte
		msg    IDResponseMsg
		err    error
	}{
		{nil, IDResponseMsg{}, ErrParsStream},
		{[]byte{}, IDResponseMsg{}, ErrParsStream},
		{[]byte{1, 2, 3, 4}, IDResponseMsg{}, ErrParsStream},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, IDResponseMsg{}, ErrParsStream},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}, IDResponseMsg{}, ErrParsStream},
		{[]byte{1, 0, 0, 0, 0, 0, 0, 0}, IDResponseMsg{ID: 1}, nil},
		{[]byte{1, 1, 0, 0, 0, 0, 0, 0}, IDResponseMsg{ID: 257}, nil},
	}

	for _, tt := range tests {
		actual, err := DeserializeIDRes(tt.stream)
		if !ChkIDResponseMsgEq(actual, tt.msg) || err != tt.err {
			t.Errorf("DeserializeIDRes: expected %d-%s, actual %d-%s", tt.msg.ID, tt.err, actual.ID, err)
		}
	}
}

func TestDeserializeListRes(t *testing.T) {
	var tests = []struct {
		stream []byte
		msg    ListResponseMsg
		err    error
	}{
		{nil, ListResponseMsg{}, nil},
		{[]byte{}, ListResponseMsg{}, nil},
		{[]byte{1, 2, 3, 4}, ListResponseMsg{}, ErrParsStream},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, ListResponseMsg{}, ErrParsStream},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7}, ListResponseMsg{}, ErrParsStream},
		{[]byte{1, 0, 0, 0, 0, 0, 0, 0}, ListResponseMsg{IDs: []uint64{1}}, nil},
		{[]byte{1, 1, 0, 0, 0, 0, 0, 0}, ListResponseMsg{IDs: []uint64{257}}, nil},
		{[]byte{1, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0}, ListResponseMsg{IDs: []uint64{1, 257}}, nil},
		{[]byte{1, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 72, 196, 124, 38, 231, 35, 0, 0}, ListResponseMsg{IDs: []uint64{1, 257, 39475690128456}}, nil},
	}

	for _, tt := range tests {
		actual, err := DeserializeListRes(tt.stream)
		if !ChkListResponseMsgEq(actual, tt.msg) || err != tt.err {
			t.Errorf("DeserializeIDRes: expected %v-%s, actual %d-%s", tt.msg.IDs, tt.err, actual.IDs, err)
		}
	}
}

func TestDeserializeRelayReq(t *testing.T) {
	var tests = []struct {
		stream []byte
		msg    RelayRequestMsg
		err    error
	}{
		{nil, RelayRequestMsg{}, ErrParsStream},
		{[]byte{}, RelayRequestMsg{}, ErrParsStream},
		{[]byte{1, 2, 3, 4}, RelayRequestMsg{}, ErrParsStream},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9}, RelayRequestMsg{}, ErrParsStream},
		{[]byte{0, 2, 3, 4, 5, 6, 7, 8, 9, 1}, RelayRequestMsg{}, ErrParsStream},
		{[]byte{2, 1, 0, 0, 0, 0, 0, 0, 0, 200}, RelayRequestMsg{}, ErrParsStream},
		{[]byte{2, 1, 0, 0, 0, 0, 0, 0, 0, 72, 196, 124, 38, 231, 35, 0, 0}, RelayRequestMsg{}, ErrParsStream},
		{[]byte{1, 1, 0, 0, 0, 0, 0, 0, 0, 200}, RelayRequestMsg{IDs: []uint64{1}, Data: []byte{200}}, nil},
		{[]byte{1, 1, 0, 0, 0, 0, 0, 0, 0, 72, 196, 124, 38, 231, 35, 0, 0, 200, 201, 202, 203, 204, 205}, RelayRequestMsg{IDs: []uint64{1}, Data: []byte{72, 196, 124, 38, 231, 35, 0, 0, 200, 201, 202, 203, 204, 205}}, nil},
		{[]byte{2, 1, 0, 0, 0, 0, 0, 0, 0, 72, 196, 124, 38, 231, 35, 0, 0, 200, 201, 202, 203, 204, 205}, RelayRequestMsg{IDs: []uint64{1, 39475690128456}, Data: []byte{200, 201, 202, 203, 204, 205}}, nil},
	}

	for _, tt := range tests {
		actual, err := DeserializeRelayReq(tt.stream)
		if !ChkRelayRequestMsgEq(actual, tt.msg) || err != tt.err {
			t.Errorf("DeserializeRelayReq: expected %v-%v-%s, actual %v-%v-%s", tt.msg.IDs, tt.msg.Data, tt.err, actual.IDs, actual.Data, err)
		}
	}
}

func TestDeserializeRelayRes(t *testing.T) {
	var tests = []struct {
		stream []byte
		msg    RelayResponseMsg
		err    error
	}{
		{nil, RelayResponseMsg{}, ErrParsStream},
		{[]byte{}, RelayResponseMsg{}, ErrParsStream},
		{[]byte{1, 2, 3, 4}, RelayResponseMsg{}, ErrParsStream},
		{[]byte{1, 0, 0, 0, 0, 0, 0, 0}, RelayResponseMsg{}, ErrParsStream},
		{[]byte{1, 0, 0, 0, 0, 0, 0, 0, 72, 196, 124, 38, 231, 35, 0, 0}, RelayResponseMsg{SenderID: 1, Data: []byte{72, 196, 124, 38, 231, 35, 0, 0}}, nil},
		{[]byte{72, 196, 124, 38, 231, 35, 0, 0, 35, 0, 0, 200, 201, 202, 203, 204, 205}, RelayResponseMsg{SenderID: 39475690128456, Data: []byte{35, 0, 0, 200, 201, 202, 203, 204, 205}}, nil},
	}

	for _, tt := range tests {
		actual, err := DeserializeRelayRes(tt.stream)
		if !ChkRelayResponseMsgEq(actual, tt.msg) || err != tt.err {
			t.Errorf("DeserializeRelayReq: expected %d-%v-%s, actual %d-%v-%s", tt.msg.SenderID, tt.msg.Data, tt.err, actual.SenderID, actual.Data, err)
		}
	}
}
