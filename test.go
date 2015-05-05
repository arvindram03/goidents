package main

import "fmt"

func fn() (d int) {
	var a, c int
	a = 5
	a, d = 5

	fmt.Println(a)
}
