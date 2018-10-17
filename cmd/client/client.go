package main

import (
	"fmt"
)

func valTrue() bool {
	fmt.Println("valTrue gets called")
	return true
}

func valFalse() bool {
	fmt.Println("valFalse gets called")
	return false
}

type vahid struct {
	a int
	b int
}

func main() {
	aaa := []byte{2, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 83, 79, 70, 83, 79, 70, 83, 79, 70, 83, 79, 70, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 83, 79, 70, 83, 79, 70, 83, 79}
	fmt.Println(len(aaa))
	// aaa := vahid{
	// 	a: 10,
	// 	b: 20,
	// }
	// log.Fatalf("%+v\n", aaa)
}
