package test

import (
	"testing"

	"github.com/megaheart/goUtils"
	"github.com/stretchr/testify/assert"
)

func TestHeapSort(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "Already sorted array",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Reverse sorted array",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Unsorted array",
			input:    []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5},
			expected: []int{1, 1, 2, 3, 3, 4, 5, 5, 5, 6, 9},
		},
		{
			name:     "Array with duplicates",
			input:    []int{4, 2, 4, 2, 4, 2},
			expected: []int{2, 2, 2, 4, 4, 4},
		},
		{
			name:     "Single element array",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "Empty array",
			input:    []int{},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			LimitTestTime(t, func(t *testing.T) {
				heap := goUtils.HeapArray{Array: tt.input}
				result := heap.HeapSort()
				assert.Equal(t, tt.expected, result)
			}, 2000)

		})
	}
}
