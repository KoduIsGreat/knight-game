package main

import "fmt"

func main() {
	var foo uint16

	fmt.Println(int(foo))
	fmt.Println(foo & 0b11111)
	fmt.Println((foo >> 8) & 0b11111)

}
