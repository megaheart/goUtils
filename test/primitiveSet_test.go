package test

import (
	"testing"
	"time"

	"github.com/megaheart/goUtils"
	"github.com/stretchr/testify/assert"
)

func LimitTestTime(t *testing.T, testFunc func(t *testing.T), msTimeout float32) {
	t.Helper()
	// Set a timeout for the test
	done := make(chan bool)
	go func() {
		testFunc(t)
		done <- true
	}()

	select {
	case <-done:
		// Test completed within the time limit
	case <-time.After(time.Duration(msTimeout) * time.Millisecond):
		t.Fatal("Test timed out")
	}
}

func TestPrimitiveSet_Add(t *testing.T) {
	tests := []struct {
		name       string
		initValues []int
		addValue   int
		expected   []int
		timeoutMS  float32
	}{
		{"ZeroLen_Add", []int{}, 106, []int{106}, 2000},

		{"Len1_Add_Last", []int{1}, 2, []int{1, 2}, 2000},
		{"Len1_Add_First", []int{2}, 1, []int{1, 2}, 2000},
		{"Len1_Add_Duplicated", []int{2}, 2, []int{2}, 2000},

		{"Len2_Add_Last", []int{1, 2}, 3, []int{1, 2, 3}, 2000},
		{"Len2_Add_First", []int{2, 3}, 1, []int{1, 2, 3}, 2000},
		{"Len2_Add_Mid", []int{1, 3}, 2, []int{1, 2, 3}, 2000},
		{"Len2_Add_Duplicated1", []int{1, 2}, 2, []int{1, 2}, 2000},
		{"Len2_Add_Duplicated2", []int{1, 2}, 1, []int{1, 2}, 2000},

		{"OddLen_Add_Last", []int{1, 2, 3, 4, 5}, 6, []int{1, 2, 3, 4, 5, 6}, 2000},
		{"OddLen_Add_First", []int{2, 3, 8, 12, 16}, 1, []int{1, 2, 3, 8, 12, 16}, 2000},
		{"OddLen_Add_Mid", []int{2, 3, 8, 12, 16}, 9, []int{2, 3, 8, 9, 12, 16}, 2000},
		{"OddLen_Add_Mid2", []int{2, 3, 8, 12, 16}, 14, []int{2, 3, 8, 12, 14, 16}, 2000},
		{"OddLen_Add_Mid3", []int{2, 3, 8, 12, 16}, 5, []int{2, 3, 5, 8, 12, 16}, 2000},
		{"OddLen_Add_DuplicatedFirst", []int{2, 3, 8, 12, 16}, 2, []int{2, 3, 8, 12, 16}, 2000},
		{"OddLen_Add_DuplicatedLast", []int{2, 3, 8, 12, 16}, 16, []int{2, 3, 8, 12, 16}, 2000},
		{"OddLen_Add_DuplicatedMid1", []int{2, 3, 8, 12, 16}, 8, []int{2, 3, 8, 12, 16}, 2000},
		{"OddLen_Add_DuplicatedMid2", []int{2, 3, 8, 12, 16}, 3, []int{2, 3, 8, 12, 16}, 2000},
		{"OddLen_Add_DuplicatedMid3", []int{2, 3, 8, 12, 16}, 12, []int{2, 3, 8, 12, 16}, 2000},

		{"EvenLen_Add_Last", []int{1, 2, 3, 4, 5, 6}, 7, []int{1, 2, 3, 4, 5, 6, 7}, 2000},
		{"EvenLen_Add_First", []int{2, 3, 8, 12, 16, 20}, 1, []int{1, 2, 3, 8, 12, 16, 20}, 2000},
		{"EvenLen_Add_Mid", []int{2, 3, 8, 12, 16, 20}, 9, []int{2, 3, 8, 9, 12, 16, 20}, 2000},
		{"EvenLen_Add_Mid2", []int{2, 3, 8, 12, 16, 20}, 14, []int{2, 3, 8, 12, 14, 16, 20}, 2000},
		{"EvenLen_Add_Mid3", []int{2, 3, 8, 12, 16, 19}, 5, []int{2, 3, 5, 8, 12, 16, 19}, 2000},
		{"EvenLen_Add_DuplicatedFirst", []int{2, 3, 8, 12, 16, 19}, 2, []int{2, 3, 8, 12, 16, 19}, 2000},
		{"EvenLen_Add_DuplicatedLast", []int{2, 3, 8, 12, 16, 19}, 19, []int{2, 3, 8, 12, 16, 19}, 2000},
		{"EvenLen_Add_DuplicatedMid1", []int{2, 3, 8, 12, 16, 19}, 8, []int{2, 3, 8, 12, 16, 19}, 2000},
		{"EvenLen_Add_DuplicatedMid2", []int{2, 3, 8, 12, 16, 19}, 3, []int{2, 3, 8, 12, 16, 19}, 2000},
		{"EvenLen_Add_DuplicatedMid3", []int{2, 3, 8, 12, 16, 19}, 12, []int{2, 3, 8, 12, 16, 19}, 2000},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			LimitTestTime(t, func(t *testing.T) {
				s := goUtils.NewPrimitiveSetWithInitValues[int](tc.initValues)
				s.Add(tc.addValue)
				actual := s.Array
				assert.Equal(t, tc.expected, actual)
			}, tc.timeoutMS)
		})
	}
}

func TestPrimitiveSet_Contains(t *testing.T) {
	tests := []struct {
		name         string
		initValues   []int
		findingValue int
		expected     bool
		timeoutMS    float32
	}{
		{"ZeroLen_NotContains", []int{}, 106, false, 2000},
		{"ZeroLen_NotContains", []int{}, -2, false, 2000},

		{"Len1_Contains", []int{1}, 1, true, 2000},
		{"Len1_NotContains", []int{2}, 1, false, 2000},
		{"Len1_NotContains2", []int{2}, 3, false, 2000},

		{"Len2_Contains_Last", []int{1, 2}, 2, true, 2000},
		{"Len2_Contains_First", []int{2, 3}, 2, true, 2000},
		{"Len2_NotContains1", []int{1, 3}, 2, false, 2000},
		{"Len2_NotContains2", []int{1, 3}, 0, false, 2000},
		{"Len2_NotContains3", []int{1, 3}, 4, false, 2000},
		{"Len2_NotContains4", []int{1, 3}, -2, false, 2000},
		{"Len2_NotContains5", []int{1, 3}, 40, false, 2000},

		{"OddLen_Contains_Last", []int{1, 2, 3, 4, 5}, 5, true, 2000},
		{"OddLen_Contains_First", []int{2, 3, 8, 12, 16}, 2, true, 2000},
		{"OddLen_Contains_Mid", []int{2, 3, 8, 12, 16}, 8, true, 2000},
		{"OddLen_Contains_Mid2", []int{2, 3, 8, 12, 16}, 12, true, 2000},
		{"OddLen_Contains_Mid3", []int{2, 3, 8, 12, 16}, 3, true, 2000},
		{"OddLen_NotContains_First", []int{2, 3, 8, 12, 16}, 1, false, 2000},
		{"OddLen_NotContains_Last", []int{2, 3, 8, 12, 16}, 17, false, 2000},
		{"OddLen_NotContains_Mid1", []int{2, 3, 8, 12, 16}, 4, false, 2000},
		{"OddLen_NotContains_Mid2", []int{2, 3, 8, 12, 16}, 15, false, 2000},
		{"OddLen_NotContains_Mid3", []int{2, 3, 8, 12, 16}, 11, false, 2000},

		{"EvenLen_Contains_Last", []int{1, 2, 3, 4, 5, 6}, 6, true, 2000},
		{"EvenLen_Contains_First", []int{2, 3, 8, 12, 16, 20}, 2, true, 2000},
		{"EvenLen_Contains_Mid", []int{2, 3, 8, 12, 16, 20}, 8, true, 2000},
		{"EvenLen_Contains_Mid2", []int{2, 3, 8, 12, 16, 20}, 3, true, 2000},
		{"EvenLen_Contains_Mid3", []int{2, 3, 8, 12, 16, 19}, 16, true, 2000},
		{"EvenLen_NotContains_First", []int{2, 3, 8, 12, 16, 19}, 1, false, 2000},
		{"EvenLen_NotContains_Last", []int{2, 3, 8, 12, 16, 19}, 20, false, 2000},
		{"EvenLen_NotContains_Mid1", []int{2, 3, 8, 12, 16, 19}, 9, false, 2000},
		{"EvenLen_NotContains_Mid2", []int{2, 3, 8, 12, 16, 19}, 5, false, 2000},
		{"EvenLen_NotContains_Mid3", []int{2, 3, 8, 12, 16, 19}, 13, false, 2000},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			LimitTestTime(t, func(t *testing.T) {
				s := goUtils.NewPrimitiveSetWithInitValues[int](tc.initValues)
				actual := s.Contains(tc.findingValue)
				assert.Equal(t, tc.expected, actual)
			}, tc.timeoutMS)
		})
	}
}

func TestPrimitiveSet_SearchIndex(t *testing.T) {
	tests := []struct {
		name         string
		initValues   []int
		findingValue int
		expected     int
		timeoutMS    float32
	}{
		{"ZeroLen_SearchIndex", []int{}, 106, -1, 2000},
		{"ZeroLen_SearchIndex", []int{}, -2, -1, 2000},

		{"Len1_SearchIndex", []int{1}, 1, 0, 2000},
		{"Len1_FailSearchIndex", []int{2}, 1, -1, 2000},
		{"Len1_FailSearchIndex2", []int{2}, 3, -1, 2000},
		{"Len1_FailSearchIndex3", []int{2}, -2, -1, 2000},
		{"Len1_FailSearchIndex4", []int{2}, 20, -1, 2000},

		{"Len2_SearchIndex_Last", []int{1, 2}, 2, 1, 2000},
		{"Len2_SearchIndex_First", []int{2, 3}, 2, 0, 2000},
		{"Len2_FailSearchIndex", []int{1, 3}, 2, -1, 2000},
		{"Len2_FailSearchIndex1", []int{1, 3}, 0, -1, 2000},
		{"Len2_FailSearchIndex2", []int{1, 3}, 4, -1, 2000},
		{"Len2_FailSearchIndex3", []int{1, 3}, -2, -1, 2000},
		{"Len2_FailSearchIndex4", []int{1, 3}, 40, -1, 2000},

		{"OddLen_SearchIndex_Last", []int{1, 2, 3, 4, 5}, 5, 4, 2000},
		{"OddLen_SearchIndex_First", []int{2, 3, 8, 12, 16}, 2, 0, 2000},
		{"OddLen_SearchIndex_Mid", []int{2, 3, 8, 12, 16}, 8, 2, 2000},
		{"OddLen_SearchIndex_Mid2", []int{2, 3, 8, 12, 16}, 12, 3, 2000},
		{"OddLen_SearchIndex_Mid3", []int{2, 3, 8, 12, 16}, 3, 1, 2000},
		{"OddLen_FailSearchIndex_First", []int{2, 3, 8, 12, 16}, -1, -1, 2000},
		{"OddLen_FailSearchIndex_First2", []int{2, 3, 8, 12, 16}, 1, -1, 2000},
		{"OddLen_FailSearchIndex_Last", []int{2, 3, 8, 12, 16}, 17, -1, 2000},
		{"OddLen_FailSearchIndex_Last2", []int{2, 3, 8, 12, 16}, 20, -1, 2000},
		{"OddLen_FailSearchIndex_Mid1", []int{2, 3, 8, 12, 16}, 4, -1, 2000},
		{"OddLen_FailSearchIndex_Mid2", []int{2, 3, 8, 12, 16}, 15, -1, 2000},
		{"OddLen_FailSearchIndex_Mid3", []int{2, 3, 8, 12, -16}, 11, -1, 2000},

		{"EvenLen_SearchIndex_Last", []int{1, 2, 3, 4, 5, 6}, 6, 5, 2000},
		{"EvenLen_SearchIndex_First", []int{2, 3, 8, 12, 16, 20}, 2, 0, 2000},
		{"EvenLen_SearchIndex_Mid", []int{2, 3, 8, 12, 16, 20}, 8, 2, 2000},
		{"EvenLen_SearchIndex_Mid2", []int{2, 3, 8, 12, 16, 20}, 3, 1, 2000},
		{"EvenLen_SearchIndex_Mid3", []int{2, 3, 8, 12, 16, 19}, 16, 4, 2000},
		{"EvenLen_FailSearchIndex_First", []int{2, 3, 8, 12, 16, 19}, -1, -1, 2000},
		{"EvenLen_FailSearchIndex_First2", []int{2, 3, 8, 12, 16, 19}, 1, -1, 2000},
		{"EvenLen_FailSearchIndex_Last", []int{2, 3, 8, 12, 16, 19}, 20, -1, 2000},
		{"EvenLen_FailSearchIndex_Last2", []int{2, 3, 8, 12, 16, 19}, 38, -1, 2000},
		{"EvenLen_FailSearchIndex_Mid1", []int{2, 3, 8, 12, 16, 19}, 9, -1, 2000},
		{"EvenLen_FailSearchIndex_Mid2", []int{2, 3, 8, 12, 16, 19}, 5, -1, 2000},
		{"EvenLen_FailSearchIndex_Mid3", []int{2, 3, 8, 12, 16, 19}, 13, -1, 2000},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			LimitTestTime(t, func(t *testing.T) {
				s := goUtils.NewPrimitiveSetWithInitValues[int](tc.initValues)
				actual := s.SearchIndex(tc.findingValue)
				assert.Equal(t, tc.expected, actual)
			}, tc.timeoutMS)
		})
	}
}
