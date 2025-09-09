package main

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/megaheart/goUtils"
)

func Random(a, b int) int {
	// Generate a random integer in [a, b)
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(b-a)))
	return int(n.Int64()) + a
}

func main() {
	s := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n := 3
	count := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	// Pick n random elements from s
	for i := 0; i < 1000000; i++ {
		picked := goUtils.PickRandom(s, n, Random)

		// fmt.Printf("%v\n", picked)
		for _, v := range picked {
			count[v-1]++
		}
	}

	for i, v := range count {
		fmt.Printf("%d: %d\n", i+1, v)
	}
}
