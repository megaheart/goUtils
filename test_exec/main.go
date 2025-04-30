package main

import (
	"fmt"

	"github.com/megaheart/goUtils"
)

func main() {
	s := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n := 3
	count := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	// Pick n random elements from s
	for i := 0; i < 1000000; i++ {
		picked := goUtils.PickRandom(s, n)

		// fmt.Printf("%v\n", picked)
		for _, v := range picked {
			count[v-1]++
		}
	}

	for i, v := range count {
		fmt.Printf("%d: %d\n", i+1, v)
	}
}
