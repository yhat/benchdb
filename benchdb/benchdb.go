package benchdb

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/tools/benchmark/parse"
)

// DB represents a sql database that can be used to write
// benchmark data to.
type BenchDB struct {
	// Test represents options for the go test command
	//Test *TestOpts
	// Regex is a regex used to pass to pattern match and run
	// benchmark sets
	Regex string
	// Driver is a database driver name
	Driver string
	// ConnStr is a sql connection string
	ConnStr string
	// CsvWriter is a csv Writer that is used to format benchmark data
	// when writing to stdout
	CsvWriter csv.Writer

	dbConn *sql.DB
}

// WriteBenchSet writes data from a parsed benchmark set to a DB table and a csv.Writer
// It returns the number of benchmark tests written and any error encountered.
func (benchdb *BenchDB) WriteBenchSet(benchset parse.Set) (int, error) {
	cnt := 0
	for _, v := range benchset {
		n := len(v)
		for i := 0; i < n; i++ {
			// TODO: Convert parsing to use data from parse.Benchmark
			// struct.
			row := strings.Split(v[i].String(), " ")
			err := benchdb.CsvWriter.Write(row)
			if err != nil {
				fmt.Println("Error:", err)
				return 0, err
			}
			cnt++
		}
	}
	return cnt, nil
}

// Run executes all of the go test benchmarks that match regexpr in the
// current directory. By default it does not run unit tests. It returns
// parsed benchmark stats in a parse.Set and returns any error encountered.
func (benchdb *BenchDB) Run() (int, error) {
	_, err := sql.Open(benchdb.Driver, benchdb.ConnStr)
	if err != nil {
		return 0, fmt.Errorf("could not connect to db: %v", err)
	}
	cmd := exec.Command("go", "test", "-bench", benchdb.Regex, "-test.run", "XXX", "-benchmem")
	var out bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &out)
	err = cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("command failed: %v", err)
	}
	benchSet, err := parse.ParseSet(&out)
	if err != nil {
		return 0, fmt.Errorf("failed to parse benchmark data: %v", err)
	}
	n, err := benchdb.WriteBenchSet(benchSet)
	if err != nil {
		return 0, fmt.Errorf("failed to write benchSet: %v", err)
	}
	return n, nil
}
