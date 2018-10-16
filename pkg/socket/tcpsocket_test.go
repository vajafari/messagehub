package socket

import (
	"fmt"
	"testing"

	"github.com/vajafari/messagehub/pkg/message"
)

func TestInspect(t *testing.T) {

	pi := packetInspector{}
	pi.resetVariables()
	msgTypeLen := make(map[byte]int)
	msgTypeLen[byte(message.IDMgsCode)] = 1
	msgTypeLen[byte(message.ListMgsCode)] = 2
	msgTypeLen[byte(message.RelayMgsCode)] = 3

	var tests = []struct {
		bb                []byte
		expectedRes       []rDataPacket
		expectedInspector packetInspector
		RequeireReset     bool
	}{
	// {[]byte{83, 79, 70, 83, 79, 70, 10},
	// 	nil,
	// 	packetInspector{
	// 		completeFindPrefix: true,
	// 		partialFindPrefix:  false,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      0,
	// 		lastIndexPrefix:    6,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },

	// {[]byte{0, 1, 2, 3, 83, 79, 70, 83, 79, 70, 10},
	// 	packetInspector{
	// 		completeFindPrefix: true,
	// 		partialFindPrefix:  false,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      0,
	// 		lastIndexPrefix:    10,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },
	// {[]byte{83, 79, 70, 83, 79, 70, 10, 0, 0, 0},
	// 	packetInspector{
	// 		completeFindPrefix: true,
	// 		partialFindPrefix:  false,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      0,
	// 		lastIndexPrefix:    6,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },

	// {[]byte{0, 1, 2, 3, 4, 5, 83, 79, 70, 83, 79, 70, 10, 0, 0, 0},
	// 	packetInspector{
	// 		completeFindPrefix: true,
	// 		partialFindPrefix:  false,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      0,
	// 		lastIndexPrefix:    12,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },

	// {[]byte{0, 1, 2, 3, 4, 5, 83, 79, 70, 83, 79, 70, 9, 0, 0, 0},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  false,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      0,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },

	// {[]byte{0, 1, 2, 3, 4, 5, 10, 79, 70, 83, 79, 70, 10, 0, 0, 0},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  false,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      0,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },
	// {[]byte{0, 1, 2, 3, 4, 5, 83, 79, 70, 83, 10, 70, 10, 0, 0, 0},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  false,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      0,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },

	// {[]byte{0, 1, 2, 3, 4, 5, 10, 79, 70, 83, 79, 70, 10, 0, 0, 0},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  false,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      0,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },

	// {[]byte{83, 79, 70, 83, 79, 70, 3, 4, 5, 83, 79, 70, 83, 79, 70, 9, 0, 0, 0},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  false,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      0,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },

	// {[]byte{5, 83, 79, 70},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  true,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      3,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },

	// {[]byte{5, 83, 79, 70},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  true,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      3,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },
	// {[]byte{5, 83, 79, 70, 83, 79, 70},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  true,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      6,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },
	// {[]byte{5, 83, 79},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  true,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      2,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, false,
	// },
	// {[]byte{70, 83, 79},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  true,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      5,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, false,
	// },
	// {[]byte{70},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  true,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      6,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, false,
	// },
	// {[]byte{10, 11, 12, 13},
	// 	packetInspector{
	// 		completeFindPrefix: true,
	// 		partialFindPrefix:  false,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      0,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, true,
	// },

	// {[]byte{5, 83, 79},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  true,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      2,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, false,
	// },
	// {[]byte{70, 83, 79},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  true,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      5,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, false,
	// },
	// {[]byte{10, 83, 79},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  true,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      2,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, false,
	// },

	// {[]byte{5, 83, 79},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  true,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      2,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, false,
	// },
	// {[]byte{70, 83, 79},
	// 	packetInspector{
	// 		completeFindPrefix: false,
	// 		partialFindPrefix:  true,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      5,
	// 		lastIndexPrefix:    0,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, false,
	// },
	// {[]byte{10, 83, 79, 70, 83, 79, 70, 10, 1, 2, 3},
	// 	packetInspector{
	// 		completeFindPrefix: true,
	// 		partialFindPrefix:  false,
	// 		headerVerified:     false,
	// 		prevPrefixCnt:      0,
	// 		lastIndexPrefix:    7,
	// 		currentPkgLen:      0,
	// 		curPkgHeader:       nil,
	// 		curPkg:             nil,
	// 	}, false,
	// },
	}

	for _, tt := range tests {
		actual := pi.inspect(tt.bb, msgTypeLen)
		if !checkEqPacketInspector(pi, tt.expectedInspector) || checkEqRData(actual, tt.expectedRes) {
			t.Error("Not expected result")
		}
		if tt.RequeireReset {
			pi.resetVariables()
		}
	}
}

func TestFindPrefix(t *testing.T) {

	pi := packetInspector{}
	pi.resetVariables()
	var tests = []struct {
		bb            []byte
		expected      packetInspector
		RequeireReset bool
	}{
		{[]byte{83, 79, 70, 83, 79, 70, 10},
			packetInspector{
				completeFindPrefix: true,
				partialFindPrefix:  false,
				headerVerified:     false,
				prevPrefixCnt:      0,
				lastIndexPrefix:    6,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},
		{[]byte{0, 1, 2, 3, 83, 79, 70, 83, 79, 70, 10},
			packetInspector{
				completeFindPrefix: true,
				partialFindPrefix:  false,
				headerVerified:     false,
				prevPrefixCnt:      0,
				lastIndexPrefix:    10,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},
		{[]byte{83, 79, 70, 83, 79, 70, 10, 0, 0, 0},
			packetInspector{
				completeFindPrefix: true,
				partialFindPrefix:  false,
				headerVerified:     false,
				prevPrefixCnt:      0,
				lastIndexPrefix:    6,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},

		{[]byte{0, 1, 2, 3, 4, 5, 83, 79, 70, 83, 79, 70, 10, 0, 0, 0},
			packetInspector{
				completeFindPrefix: true,
				partialFindPrefix:  false,
				headerVerified:     false,
				prevPrefixCnt:      0,
				lastIndexPrefix:    12,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},

		{[]byte{0, 1, 2, 3, 4, 5, 83, 79, 70, 83, 79, 70, 9, 0, 0, 0},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  false,
				headerVerified:     false,
				prevPrefixCnt:      0,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},

		{[]byte{0, 1, 2, 3, 4, 5, 10, 79, 70, 83, 79, 70, 10, 0, 0, 0},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  false,
				headerVerified:     false,
				prevPrefixCnt:      0,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},
		{[]byte{0, 1, 2, 3, 4, 5, 83, 79, 70, 83, 10, 70, 10, 0, 0, 0},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  false,
				headerVerified:     false,
				prevPrefixCnt:      0,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},

		{[]byte{0, 1, 2, 3, 4, 5, 10, 79, 70, 83, 79, 70, 10, 0, 0, 0},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  false,
				headerVerified:     false,
				prevPrefixCnt:      0,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},

		{[]byte{83, 79, 70, 83, 79, 70, 3, 4, 5, 83, 79, 70, 83, 79, 70, 9, 0, 0, 0},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  false,
				headerVerified:     false,
				prevPrefixCnt:      0,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},

		{[]byte{5, 83, 79, 70},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  true,
				headerVerified:     false,
				prevPrefixCnt:      3,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},

		{[]byte{5, 83, 79, 70},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  true,
				headerVerified:     false,
				prevPrefixCnt:      3,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},
		{[]byte{5, 83, 79, 70, 83, 79, 70},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  true,
				headerVerified:     false,
				prevPrefixCnt:      6,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},
		{[]byte{5, 83, 79},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  true,
				headerVerified:     false,
				prevPrefixCnt:      2,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, false,
		},
		{[]byte{70, 83, 79},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  true,
				headerVerified:     false,
				prevPrefixCnt:      5,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, false,
		},
		{[]byte{70},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  true,
				headerVerified:     false,
				prevPrefixCnt:      6,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, false,
		},
		{[]byte{10, 11, 12, 13},
			packetInspector{
				completeFindPrefix: true,
				partialFindPrefix:  false,
				headerVerified:     false,
				prevPrefixCnt:      0,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, true,
		},

		{[]byte{5, 83, 79},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  true,
				headerVerified:     false,
				prevPrefixCnt:      2,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, false,
		},
		{[]byte{70, 83, 79},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  true,
				headerVerified:     false,
				prevPrefixCnt:      5,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, false,
		},
		{[]byte{10, 83, 79},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  true,
				headerVerified:     false,
				prevPrefixCnt:      2,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, false,
		},

		{[]byte{5, 83, 79},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  true,
				headerVerified:     false,
				prevPrefixCnt:      2,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, false,
		},
		{[]byte{70, 83, 79},
			packetInspector{
				completeFindPrefix: false,
				partialFindPrefix:  true,
				headerVerified:     false,
				prevPrefixCnt:      5,
				lastIndexPrefix:    0,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, false,
		},
		{[]byte{10, 83, 79, 70, 83, 79, 70, 10, 1, 2, 3},
			packetInspector{
				completeFindPrefix: true,
				partialFindPrefix:  false,
				headerVerified:     false,
				prevPrefixCnt:      0,
				lastIndexPrefix:    7,
				currentPkgLen:      0,
				curPkgHeader:       nil,
				curPkg:             nil,
			}, false,
		},
	}

	for _, tt := range tests {
		fmt.Printf("%+v\n", pi)
		pi.findPrefix(tt.bb)
		if !checkEqPacketInspector(pi, tt.expected) {
			t.Errorf("expected %+v, actual %+v", tt.expected, pi)
		}
		if tt.RequeireReset {
			pi.resetVariables()
		}
	}
}

func checkEqPacketInspector(a, b packetInspector) bool {
	return a.completeFindPrefix == b.completeFindPrefix &&
		a.currentPkgLen == b.currentPkgLen &&
		a.headerVerified == b.headerVerified &&
		a.lastIndexPrefix == b.lastIndexPrefix &&
		a.partialFindPrefix == b.partialFindPrefix &&
		a.prevPrefixCnt == b.prevPrefixCnt &&
		checkEqByte(a.curPkg, b.curPkg) &&
		checkEqByte(a.curPkgHeader, b.curPkgHeader)

}

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

func checkEqRData(a, b []rDataPacket) bool {

	if len(a) != len(b) {
		return false
	}
	if len(a) > 0 {
		for i := 0; i < len(a); i++ {
			if !checkEqByte(a[i].data, b[i].data) || a[i].typ != b[i].typ {
				return false
			}
		}
	}
	return true
}
