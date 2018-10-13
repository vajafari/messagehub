package main

import (
	"encoding/binary"
	"fmt"
)

type MessageHeader struct {
	FrmStart    [10]byte
	CommandType byte
	Len         uint
}

func main() {

	var bbbb []byte
	fmt.Println(len(bbbb))
	bb := make([]byte, 8)
	binary.LittleEndian.PutUint64(bb, uint64(4739))
	slc := []uint64{39475690128456}

	for _, n := range slc {
		binary.LittleEndian.PutUint64(bb, uint64(n))
		fmt.Print(bb[0])
		fmt.Print(", ")
		fmt.Print(bb[1])
		fmt.Print(", ")
		fmt.Print(bb[2])
		fmt.Print(", ")
		fmt.Print(bb[3])
		fmt.Print(", ")
		fmt.Print(bb[4])
		fmt.Print(", ")
		fmt.Print(bb[5])
		fmt.Print(", ")
		fmt.Print(bb[6])
		fmt.Print(", ")
		fmt.Print(bb[7])
		fmt.Print(", ")
	}

}
