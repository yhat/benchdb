package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	"github.com/yhat/gobenchdb/benchdb"
)

func main() {
	// regex for benchmark suite
	suite := flag.String("-test.bench", ".", "benchmark regexp")
	flag.Parse()

	// write data in csv format to stdout
	// TODO: convert this to write to a sql db
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()
	_, err := (&benchdb.BenchDB{
		Regex:     *suite,
		Driver:    "mysql",
		ConnStr:   "mysql://mysql@passs/fooDB",
		CsvWriter: *writer,
	}).Run()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(2)
	}
}
