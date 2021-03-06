/*
benchdb is a command line tool for running and storing go benchmark data in a
database. It runs the go test bench command in the current working directory
and parses the output using the parse package. The parsed data is then written to
sql database of your choice.

   Usage:
       benchdb [options...]

       Options:
             -conn        postgres database connection string
             -table       postgres table name
             -test.bench  run only those benchmarks matching the regular expression

benchdb assumes a schema is created in a database

    CREATE TABLE IF NOT EXISTS benchmarks (
        id                    serial primary key,
        batch_id              varchar(50),
        latest_sha            varchar(50),
        datetime              timestamp without time zone,
        name                  varchar(50),
        n                     integer,
        ns_op                 double precision,
        allocated_bytes_op    integer,
        allocs_op             integer
    );

*/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yhat/benchdb/benchdb"
)

var (
	conn      = flag.String("conn", "", "postgres database connection string")
	table     = flag.String("table", "", "postgres table name")
	testBench = flag.String("test.bench", ".",
		"run only those benchmarks matching the regular expression")
)

const Postgres = "postgres"

var nsha = 7

var usage = `Usage: benchdb [options...]

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

	// Initalize a BenchPSQL to and run benchmarks.
	err := (&benchdb.BenchPSQL{
		Config: &benchdb.BenchDBConfig{
			Regex:  tregex,
			ShaLen: nsha,
		},
		Driver:    Postgres,
		ConnStr:   c,
		TableName: t,
	}).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
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
