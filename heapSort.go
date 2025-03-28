package goUtils

import "fmt"

type HeapArray struct {
	Array []int
}

func (heap *HeapArray) BuildHeap() {
	for i := len(heap.Array)/2 - 1; i >= 0; i-- {
		heap.Heapify(len(heap.Array), i)
	}
}

func (heap *HeapArray) Heapify(n, i int) {
	largest := i
	left := 2*i + 1
	right := 2*i + 2

	if left < n && heap.Array[left] > heap.Array[largest] {
		largest = left
	}

	if right < n && heap.Array[right] > heap.Array[largest] {
		largest = right
	}

	if largest != i {
		heap.Array[i], heap.Array[largest] = heap.Array[largest], heap.Array[i]
		heap.Heapify(n, largest)
	}
}

func (heap *HeapArray) HeapSort() []int {
	heap.BuildHeap()
	// fmt.Print("BuildHeap: ")
	heap.Print()

	for i := len(heap.Array) - 1; i > 0; i-- {
		heap.Array[0], heap.Array[i] = heap.Array[i], heap.Array[0]
		heap.Heapify(i, 0)

		// fmt.Printf("Heapify %d: ", i)
		heap.Print()
	}

	return heap.Array
}

func (heap *HeapArray) Print() {
	str := "["
	for i, v := range heap.Array {
		str += fmt.Sprintf("%v", v)
		if i != len(heap.Array)-1 {
			str += ", "
		}
	}
	str += "]"
	// fmt.Println(str)
}

// func main() {
// 	arr := []int{12, 11, 13, 5, 6, 7}
// 	heap := HeapArray{Array: arr}
// 	heap.Print()
// 	heap.HeapSort()
// 	fmt.Print("Sorted: ")
// 	heap.Print()
// }
