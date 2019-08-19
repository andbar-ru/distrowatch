package main

import (
	"fmt"

	"github.com/andbar-ru/distrowatch/show"
)

func check(err error) {
	panic(err)
}

func main() {
	var coords, err = show.GetCoords()
	check(err)
	fmt.Println(coords)
	var distrs, err = show.GetDistrs()
	check(err)
	fmt.Println(distrs)
}
