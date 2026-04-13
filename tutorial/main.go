package main

import "fmt"

func modifySlice(s [3]int) {
	s[0] = 100
	fmt.Println(s)
}

func main() {
	s := [3]int{1, 2, 3}
	modifySlice(s)
	fmt.Println(s) // [100 2 3]
}
