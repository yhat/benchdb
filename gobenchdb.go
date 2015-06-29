package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	"github.com/yhat/gobenchdb/benchdb"
)

func main() {
	// Parse command line args
	suite := flag.String("test.bench", ".", "benchmark regexp")
	connStr := flag.String("conn", "", "postgres database connection string")
	dbtable := flag.String("table", "", "postgres table")
	flag.Parse()

	// write data in csv format to stdout
	// TODO: convert this to write to a sql db
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()
	_, err := (&benchdb.BenchDB{
		Regex:     *suite,
		Driver:    "postgres",
		ConnStr:   *connStr,
		TableName: *dbtable,
		CsvWriter: *writer,
	}).Run()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(2)
	}
}
