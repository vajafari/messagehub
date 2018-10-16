package main

import (
	"fmt"
	"log"
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
	aaa := vahid{
		a: 10,
		b: 20,
	}
	log.Fatalf("%+v\n", aaa)
}
