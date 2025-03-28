package main

import (
	"fmt"

	"github.com/megaheart/goUtils"
)

func main() {
	s := goUtils.NewPrimitiveSetWithInitValues[int](
		[]int{},
	)

	s.Add(2)

	fmt.Print(s.Array)
	fmt.Print("\n")
}
