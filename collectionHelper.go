package goUtils

// import "crypto/rand"
// import (
// 	"math/rand/v2"
// )

// RandomFunc is a function type that return a random integer in [a, b) (inclusive a, exclusive b)
type RandomFunc func(a, b int) int

func PickRandom[T any](slice []T, n int, random RandomFunc) []T {
	N := len(slice)
	n = min(N, n)
	if n < 1 {
		return []T{}
	}
	clone := make([]T, N)
	copy(clone, slice)
	for i := 0; i < n; i++ {
		j := random(i, N)
		clone[i], clone[j] = clone[j], clone[i]
	}

	return clone[:n]
}
