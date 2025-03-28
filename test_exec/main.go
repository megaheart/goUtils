package main

import (
	"fmt"

	"github.com/megaheart/goUtils"
)

func main() {
	s := goUtils.NewPrimitiveSetWithInitValues[int](
		[]int{1, 2, 3, 4, 5},
	)

	s.Add(2)

	fmt.Print(s.Array)
	fmt.Print("\n")
}
