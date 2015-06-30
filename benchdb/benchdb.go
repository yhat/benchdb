package benchdb

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"golang.org/x/tools/benchmark/parse"
)

// A BenchDB manges the execution of benchmark tests using go test and
// writing a parse.Set to a database.
//
// Implementations of BenchDB for different databases are allowed by way
// of the WriteSet method.
type BenchDB interface {
	// Run executes go test bench for benchmarks matching regex in the current
	// directory. By default it does not run unit tests by way of setting test.run
	// to XXX in the call to go test. It also parses the benchSet and calls WriteSet
	// to write the benchmark data to a database. It returns any error encountered.
	Run(regex string) error

	// WriteSet is responsible for opening a postgres database connection and writing
	// a parsed benchSet to a db table. It closes the connection, returns the number of
	// benchmark tests written, and any error encountered.
	WriteSet(parse.Set) (int, error)
}

// BenchPSQL represents a BenchDB that writes benchmarks to a postgres
// database.
type BenchPSQL struct {
	Driver    string // database driver name
	ConnStr   string // sql connection string
	TableName string // database table name

	dbConn *sql.DB
}

// Run runs go test benchmarks matching regex in the current directory and writes
// benchmark data to a PSQL database by calling WriteSet. It returns any error
// encountered.
func (benchdb *BenchPSQL) Run(regex string) error {
	// Exec a subprocess for go test bench and write
	// to both stdout and a byte buffer.
	cmd := exec.Command("go", "test", "-bench", regex,
		"-test.run", "XXX", "-benchmem")
	var out bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &out)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s", out.String())
		return fmt.Errorf("command failed: %v", err)
	}

	benchSet, err := parse.ParseSet(&out)
	if err != nil {
		return fmt.Errorf("failed to parse benchmark data: %v", err)
	}

	// Writes parse set to sql database.
	_, err = benchdb.WriteSet(benchSet)
	if err != nil {
		return fmt.Errorf("failed to write benchSet to db: %v", err)
	}
	return nil
}

// WriteSet is responsible for opening a postgres database connection and writing
// a parsed benchSet to a db table. It closes the connection, returns the number of
// benchmark tests written, and any error encountered.
//
// A new sql transaction is created and committed per Benchmark in benchSet. This way if a
// db failure occurs all data from the benchSet is not lost.
func (benchdb *BenchPSQL) WriteSet(benchSet parse.Set) (int, error) {
	sqlDB, err := sql.Open(benchdb.Driver, benchdb.ConnStr)
	if err != nil {
		return 0, fmt.Errorf("could not connect to db: %v", err)
	}
	defer sqlDB.Close()
	benchdb.dbConn = sqlDB

	cnt := 0
	for _, b := range benchSet {
		n := len(b)
		for i := 0; i < n; i++ {
			val := b[i]
			err := saveBenchmark(benchdb.dbConn, benchdb.TableName, *val)
			if err != nil {
				return 0, fmt.Errorf("failed to save benchmark: %v", err)
			}
			cnt++
		}
	}
	return cnt, nil
}

func saveBenchmark(dbConn *sql.DB, table string, b parse.Benchmark) error {
	// Create a transaction per Benchmark.
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := fmt.Sprintf(`
        INSERT INTO %s
        (datetime, name, n, ns_op, allocated_bytes_op, allocs_op)
        VALUES
        ($1, $2, $3, $4, $5, $6)
        `, table)

	// Strips of leading Benchmark string in Benchmark.Name
	name := strings.TrimPrefix(strings.TrimSpace(b.Name), "Benchmark")
	ts := time.Now().UTC()

	_, err = tx.Exec(q,
		ts, name, b.N, b.NsPerOp, b.AllocedBytesPerOp, b.AllocsPerOp)
	if err != nil {
		return err
	}
	return tx.Commit()
}
