package benchdb

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

func TestRunNoDBConn(t *testing.T) {
	c := "postgres://foo@localhost:5432/benchmark"
	table := "footable"
	tregex := "BenchmarkMySort1K"
	nsha := 7

	err := (&BenchPSQL{
		Config: &BenchDBConfig{
			Regex:  tregex,
			ShaLen: nsha,
		},
		Driver:    "postgres",
		ConnStr:   c,
		TableName: table,
	}).Run()
	if err == nil {
		fmt.Println(err)
		t.Errorf("expected failure due to no db connection")
	}
}

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
