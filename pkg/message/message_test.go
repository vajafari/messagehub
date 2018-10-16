package message

import (
	"testing"
)

type messager interface {
	Type() byte
	Data() ([]byte, error)
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

func TestData(t *testing.T) {

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
	tooMuchDataRelay.Body = make([]byte, RelayMaxBodySize*2)
	for i := 0; i < RelayMaxBodySize*2; i++ {
		tooMuchDataRelay.Body[i] = 123
	}

	tooMuchDataRelayResp := RelayResponseMsg{}
	tooMuchDataRelayResp.Body = make([]byte, RelayMaxBodySize*2)
	for i := 0; i < RelayMaxBodySize*2; i++ {
		tooMuchDataRelayResp.Body[i] = 123
	}

	var tests = []struct {
		msg           messager // input
		typeName      string
		expected      []byte // expected result
		expectedError error
	}{
		{&IDRequestMsg{}, "IDRequestMsg", nil, nil},
		{&IDResponseMsg{ID: 256}, "IDResponseMsg", []byte{0, 1, 0, 0, 0, 0, 0, 0}, nil},

		{&ListRequestMsg{}, "ListRequestMsg", nil, nil},

		{&ListResponseMsg{}, "ListResponseMsg", nil, nil},
		{&ListResponseMsg{IDs: []uint64{}}, "ListResponseMsg", nil, nil},
		{&ListResponseMsg{IDs: []uint64{1}}, "ListResponseMsg", []byte{1, 0, 0, 0, 0, 0, 0, 0}, nil},
		{&ListResponseMsg{IDs: []uint64{1, 2, 3}}, "ListResponseMsg", []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0}, nil},

		{&RelayRequestMsg{}, "RelayRequestMsg", nil, ErrInvalidData},
		{&RelayRequestMsg{Body: []byte{}}, "RelayRequestMsg", nil, ErrInvalidData},
		{&RelayRequestMsg{IDs: []uint64{}}, "RelayRequestMsg", nil, ErrInvalidData},
		{&RelayRequestMsg{Body: []byte{}, IDs: []uint64{}}, "RelayRequestMsg", nil, ErrInvalidData},
		{&RelayRequestMsg{Body: []byte{123, 124, 125}}, "RelayRequestMsg", nil, ErrInvalidData},      // only data with no reciever
		{&RelayRequestMsg{IDs: []uint64{0, 1, 2, 3, 4, 54}}, "RelayRequestMsg", nil, ErrInvalidData}, // only reciever with no data
		{&tooMuchIDRelay, "RelayRequestMsg", nil, ErrInvalidData},
		{&tooMuchDataRelay, "RelayRequestMsg", nil, ErrInvalidData},
		{&RelayRequestMsg{IDs: []uint64{1, 2}, Body: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}}, "RelayRequestMsg", []byte{2, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, nil},

		{&RelayResponseMsg{}, "RelayResponseMsg", nil, ErrInvalidData},
		{&RelayResponseMsg{SenderID: 1, Body: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}}, "RelayRequestMsg", []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, nil},
		{&tooMuchDataRelayResp, "RelayResponseMsg", nil, ErrInvalidData},
	}

	for _, tt := range tests {
		actual, errActual := tt.msg.Data()
		if !checkEqByte(actual, tt.expected) || errActual != tt.expectedError {
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
		{[]byte{1, 1, 0, 0, 0, 0, 0, 0, 0, 200}, RelayRequestMsg{IDs: []uint64{1}, Body: []byte{200}}, nil},
		{[]byte{1, 1, 0, 0, 0, 0, 0, 0, 0, 72, 196, 124, 38, 231, 35, 0, 0, 200, 201, 202, 203, 204, 205}, RelayRequestMsg{IDs: []uint64{1}, Body: []byte{72, 196, 124, 38, 231, 35, 0, 0, 200, 201, 202, 203, 204, 205}}, nil},
		{[]byte{2, 1, 0, 0, 0, 0, 0, 0, 0, 72, 196, 124, 38, 231, 35, 0, 0, 200, 201, 202, 203, 204, 205}, RelayRequestMsg{IDs: []uint64{1, 39475690128456}, Body: []byte{200, 201, 202, 203, 204, 205}}, nil},
	}

	for _, tt := range tests {
		actual, err := DeserializeRelayReq(tt.stream)
		if !ChkRelayRequestMsgEq(actual, tt.msg) || err != tt.err {
			t.Errorf("DeserializeRelayReq: expected %v-%v-%s, actual %v-%v-%s", tt.msg.IDs, tt.msg.Body, tt.err, actual.IDs, actual.Body, err)
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
		{[]byte{1, 0, 0, 0, 0, 0, 0, 0, 72, 196, 124, 38, 231, 35, 0, 0}, RelayResponseMsg{SenderID: 1, Body: []byte{72, 196, 124, 38, 231, 35, 0, 0}}, nil},
		{[]byte{72, 196, 124, 38, 231, 35, 0, 0, 35, 0, 0, 200, 201, 202, 203, 204, 205}, RelayResponseMsg{SenderID: 39475690128456, Body: []byte{35, 0, 0, 200, 201, 202, 203, 204, 205}}, nil},
	}

	for _, tt := range tests {
		actual, err := DeserializeRelayRes(tt.stream)
		if !ChkRelayResponseMsgEq(actual, tt.msg) || err != tt.err {
			t.Errorf("DeserializeRelayReq: expected %d-%v-%s, actual %d-%v-%s", tt.msg.SenderID, tt.msg.Body, tt.err, actual.SenderID, actual.Body, err)
		}
	}
}
