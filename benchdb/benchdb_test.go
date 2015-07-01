package benchdb

import (
	"math/rand"
	"sort"
	"testing"
)

func mySort(data sort.Interface, a, b int) {
	sort.Sort(data)
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
