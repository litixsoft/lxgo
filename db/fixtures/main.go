package main

import "fmt"

type M map[string]interface{}
type A []M

type helper M
type helpers A

func main() {

	chk1 := M{"key": "value"}
	chk2 := helper{"key": "value"}

	fmt.Printf("%T\n", chk1)
	fmt.Printf("%T\n", chk2)

	var trp1 interface{}
	var trp2 interface{}
	trp1 = chk1
	trp2 = chk2

	_, ok := trp1.(M)
	fmt.Println(ok)

	_, ok1 := trp2.(helpers)
	fmt.Println(ok1)

}
