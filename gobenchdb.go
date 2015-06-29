package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	"github.com/yhat/gobenchdb/benchdb"
)

var (
	conn      = flag.String("conn", "", "postgres database connection string")
	table     = flag.String("table", "", "postgres table name")
	testBench = flag.String("test.bench", ".",
		"run only those benchmarks matching the regular expression")
)

var usage = `Usage: gobenchdb [options...]

Options:
  -conn        postgres database connection string
  -table       postgres table name
  -test.bench  run only those benchmarks matching the regular expression`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

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
		flag.Usage()
		os.Exit(2)
	}
}
