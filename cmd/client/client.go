package main

import (
	"fmt"

	"github.com/vajafari/messagehub/pkg/message"
)

func main() {
	aaa, err := message.DeserializeRelayReq([]byte{3, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7})
	fmt.Println(aaa)
	fmt.Println(err)
	// a := message.RelayRequestMsg{
	// 	IDs:  []uint64{2, 4},
	// 	Data: []byte{1, 2, 3, 4, 5, 6, 7},
	// }
	// aBytes := a.Serialize()
	// fmt.Print(aBytes[5:])

	// var bbbb []byte
	// fmt.Println(len(bbbb))
	// bb := make([]byte, 8)
	// //bb := make([]byte, 8)
	// //cc := bb[4:]
	// bb[4] = 4
	// bb[5] = 5
	// bb[6] = 6
	// bb[7] = 7
	// //cc[3] = 10
	// //	bb[4] = 11
	// fmt.Println(bb)
	// bb = append(bb, 20)

	// fmt.Println(bb)
	//bb[5] = 12
	//cc = append(cc, 30)

	//cc[0] = 11
	//fmt.Println(bb)
	//fmt.Println(cc)

	// binary.LittleEndian.PutUint64(bb, uint64(4739))
	// slc := []uint64{39475690128456}

	// for _, n := range slc {
	// 	binary.LittleEndian.PutUint64(bb, uint64(n))
	// 	fmt.Print(bb[0])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[1])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[2])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[3])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[4])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[5])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[6])
	// 	fmt.Print(", ")
	// 	fmt.Print(bb[7])
	// 	fmt.Print(", ")
	// }

}
