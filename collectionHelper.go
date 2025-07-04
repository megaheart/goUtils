package goUtils

// import "crypto/rand"
import (
	"math/rand"
)

func PickRandom[T any](slice []T, n int) []T {
	N := len(slice)
	n = min(N, n)
	if n < 1 {
		return []T{}
	}
	clone := make([]T, N)
	copy(clone, slice)
	for i := 0; i < n; i++ {
		j := rand.Intn(N-i) + i
		clone[i], clone[j] = clone[j], clone[i]
	}

	return clone[:n]
}

func PickRandomWithSeed[T any](slice []T, n int, seed int64) []T {
	N := len(slice)
	n = min(N, n)
	if n < 1 {
		return []T{}
	}
	clone := make([]T, N)
	copy(clone, slice)
	random := rand.New(rand.NewSource(seed))
	for i := 0; i < n; i++ {
		j := random.Intn(N-i) + i
		clone[i], clone[j] = clone[j], clone[i]
	}

	return clone[:n]
}
