package goUtils

// import "crypto/rand"
import (
	"math/rand"
	"time"
)

func PickRandom[T any](slice []T, n int) []T {
	N := len(slice)
	n = min(N, n)
	if n < 1 {
		return []T{}
	}
	clone := make([]T, N)
	copy(clone, slice)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		j := random.Intn(N-i) + i
		clone[i], clone[j] = clone[j], clone[i]
	}

	return clone[:n]
}
