package goUtils

import (
	"slices"
	"sync"
)

type IComparable interface {
	~int | ~string | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

type PrimitiveSet[T IComparable] struct {
	Array []T
	mutex sync.RWMutex
}

// SearchIndex returns the index of the target in the sorted array of the set.
// If the target is not found, the second return value is false.
//
// The time complexity is O(log(n)), where n is the length of the array.
//
// Parameters:
//
//   - target: the value to search for
//
// Returns:
//
//   - int: the index of the target in the array, or -1 if the target is not found
func (s *PrimitiveSet[T]) SearchIndex(value T) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	// implement binary search
	a, b := 0, len(s.Array)-1
	mid := (a + b) / 2
	for a <= b {
		if s.Array[mid] == value {
			return mid
		} else if s.Array[mid] < value {
			a = mid + 1
		} else {
			b = mid - 1
		}
		mid = (a + b) / 2
	}
	return -1
}

func (s *PrimitiveSet[T]) Contains(value T) bool {
	return s.SearchIndex(value) != -1
}

// Add the value to the set.
// If the value is found, it returns false.
//
// The time complexity is O(log(n)), where n is the length of the array.
//
// Parameters:
//
//   - value: the value to add
//
// Returns:
//
//   - bool: true if the value is added, false if the value is already in the set
func (s *PrimitiveSet[T]) Add(value T) bool {
	// if s.Contains(value) {
	// 	return false
	// }
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// binary search to find the position to insert
	a, b := 0, len(s.Array)-1
	mid := (a + b) / 2
	for a < b {
		if s.Array[mid] < value {
			a = mid + 1
		} else {
			b = mid
		}
		mid = (a + b) / 2
	}
	if s.Array[a] == value {
		return false
	} else if s.Array[a] < value {
		a++
	}
	// insert value at position a
	s.Array = slices.Insert(s.Array, a, value)
	return true
}

// Remove the value from the set.
// If the value is not found, it returns false.
//
// The time complexity is O(log(n)), where n is the length of the array.
//
// Parameters:
//
//   - value: the value to remove
//
// Returns:
//
//   - bool: true if the value is removed, false if the value is not in the set
func (s *PrimitiveSet[T]) Remove(value T) bool {
	index := s.SearchIndex(value)
	if index == -1 {
		return false
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Array = slices.Delete(s.Array, index, index+1)
	return true
}

// Remove the value from the set with the index.
// If the index is out of range, it returns false.
//
// The time complexity is O(n), where n is the length of the array.
//
// Parameters:
//
//   - index: the index of the value to remove
//
// Returns:
//
//   - bool: true if the value is removed, false if the index is out of range
func (s *PrimitiveSet[T]) RemoveIndex(index int) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Array = slices.Delete(s.Array, index, index+1)
	return true
}

// Initialize the set with an empty array.
func NewPrimitiveSet[T IComparable]() *PrimitiveSet[T] {
	s := new(PrimitiveSet[T])
	s.Array = make([]T, 0)
	return s
}

// Initialize the set with an empty array.
func NewPrimitiveSetWithCapacity[T IComparable](capacity int) *PrimitiveSet[T] {
	s := new(PrimitiveSet[T])
	s.Array = make([]T, 0, capacity)
	return s
}

// Initialize the set with an array of values.
func NewPrimitiveSetWithInitValues[T IComparable](initValues []T) *PrimitiveSet[T] {
	s := new(PrimitiveSet[T])
	s.Array = make([]T, len(initValues))
	copy(s.Array, initValues)
	slices.Sort(s.Array)
	return s
}
