package goUtils

import (
	"cmp"
	"slices"
)

type PrimitiveReadOnlyMap[K IComparable, V any] struct {
	key   *PrimitiveSet[K]
	value []V
}

func (p *PrimitiveReadOnlyMap[K, V]) Get(key K) (V, bool) {
	index := p.key.SearchIndex(key)
	if index == -1 {
		var zeroValue V
		return zeroValue, false
	}
	return p.value[index], true
}

func (p *PrimitiveReadOnlyMap[K, V]) Contains(key K) bool {
	return p.key.Contains(key)
}

func (p *PrimitiveReadOnlyMap[K, V]) Keys() []K {
	return p.key.Array
}

func (p *PrimitiveReadOnlyMap[K, V]) Values() []V {
	return p.value
}

func (p *PrimitiveReadOnlyMap[K, V]) Len() int {
	return len(p.key.Array)
}

// Initialize a new PrimitiveReadOnlyMap with the given key-value pairs.
func NewPrimitiveReadOnlyMap[K IComparable, V any](initMap map[K]V) *PrimitiveReadOnlyMap[K, V] {
	keyValuePair := make([]Tuple2[K, V], 0, len(initMap))
	for k, v := range initMap {
		keyValuePair = append(keyValuePair, Tuple2[K, V]{First: k, Second: v})
	}

	slices.SortFunc(keyValuePair, func(a Tuple2[K, V], b Tuple2[K, V]) int {
		return cmp.Compare(a.First, b.First)
	})

	key := NewPrimitiveSet[K]()
	key.Array = make([]K, 0, len(initMap))
	value := make([]V, 0, len(initMap))

	for _, pair := range keyValuePair {
		key.Array = append(key.Array, pair.First)
		value = append(value, pair.Second)
	}

	return &PrimitiveReadOnlyMap[K, V]{
		key:   key,
		value: value,
	}
}
