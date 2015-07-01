package benchdb

import (
	"math/rand"
	"sort"
	"testing"
)

func mySort(data sort.Interface, a, b int) {
	// Insertion sort borrowed from the std library.
	for i := a + 1; i < b; i++ {
		for j := i; j > a && data.Less(j, j-1); j-- {
			data.Swap(j, j-1)
		}
	}
}

func BenchmarkMySort1K(b *testing.B) {
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		data := make([]int, 1000)
		for i := 0; i < len(data); i++ {
			data[i] = rand.Int()
		}
		b.StartTimer()
		mySort(sort.IntSlice(data), 0, len(data))
		b.StopTimer()
	}
}
