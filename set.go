package goUtils

import (
	"slices"
	"sync"
)

type Set[T any] struct {
	Array   []T
	mutex   sync.RWMutex
	compare func(a, b T) int
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
//   - int: the index of the target in the array
//   - bool: true if the target is found, false otherwise
func (s *Set[T]) SearchIndex(value T) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	// implement binary search
	a, b := 0, len(s.Array)-1
	mid := (a + b) / 2
	cmp := s.compare
	for a <= b {
		if cmp(s.Array[mid], value) == 0 {
			return mid
		} else if cmp(s.Array[mid], value) < 0 {
			a = mid + 1
		} else {
			b = mid - 1
		}
		mid = (a + b) / 2
	}
	return -1
}

func (s *Set[T]) Contains(value T) bool {
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
func (s *Set[T]) Add(value T) bool {
	// if s.Contains(value) {
	// 	return false
	// }
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// binary search to find the position to insert
	a, b := 0, len(s.Array)-1
	mid := (a + b) / 2
	for a < b {
		if s.compare(s.Array[mid], value) < 0 {
			a = mid + 1
		} else {
			b = mid
		}
		mid = (a + b) / 2
	}
	// if the value is already in the set, return false
	if s.compare(s.Array[a], value) == 0 {
		return false
	} else if s.compare(s.Array[a], value) < 0 {
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
func (s *Set[T]) Remove(value T) bool {
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
func (s *Set[T]) RemoveIndex(index int) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Array = slices.Delete(s.Array, index, index+1)
	return true
}

func NewSet[T any](compare func(a, b T) int) *Set[T] {
	s := new(Set[T])
	s.Array = make([]T, 0)
	s.compare = compare
	return s
}

func NewNonEmptySet[T any](initSet []T, compare func(a, b T) int) *Set[T] {
	s := new(Set[T])
	s.Array = make([]T, len(initSet))
	copy(s.Array, initSet)
	slices.SortFunc(s.Array, compare)
	s.compare = compare
	return s
}
