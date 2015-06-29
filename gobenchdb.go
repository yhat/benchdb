package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	"github.com/yhat/gobenchdb/benchdb"
)

var (
	testBench = flag.String("test.bench", ".", "benchmark regexp")
	conn      = flag.String("conn", "", "postgres database connection string")
	table     = flag.String("table", "", "postgres table name")
)

func main() {
	// Parse command line args
	flag.Parse()

	// write data in csv format to stdout
	// TODO: convert this to write to a sql db
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()
	_, err := (&benchdb.BenchDB{
		Regex:     *testBench,
		Driver:    "postgres",
		ConnStr:   *conn,
		TableName: *table,
		CsvWriter: *writer,
	}).Run()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(2)
	}
}
