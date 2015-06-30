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

const Postgres = "postgres"

var usage = `Usage: gobenchdb [options...]

Options:
  -conn        postgres database connection string
  -table       postgres table name
  -test.bench  run only those benchmarks matching the regular expression`

func main() {
	// Parse command line args
	flag.Parse()
	c := *conn
	t := *table
	tregex := *testBench

	if c == "" {
		UsageExit("database conn must be specified")

	}
	if t == "" {
		UsageExit("database table must be specified")
	}

	// write data in csv format to stdout
	// TODO: convert this to write to a sql db
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()
	_, err := (&benchdb.BenchDB{
		Regex:     tregex,
		Driver:    Postgres,
		ConnStr:   c,
		TableName: t,
		CsvWriter: *writer,
	}).Run()
	if err != nil {
		UsageExit(err.Error())
	}
}

func UsageExit(message string) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n")
	}
	fmt.Fprintf(os.Stderr, usage)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}
