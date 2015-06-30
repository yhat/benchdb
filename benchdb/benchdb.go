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

// BenchDB represents a postgresql database that can be used to write
// benchmark data to.
type BenchDB struct {
	Regex     string // regex used to run benchmark sets
	Driver    string // database driver name
	ConnStr   string // sql connection string
	TableName string // database table name

	dbConn *sql.DB
}

// WriteBenchSet is responsible for opening a postgres database connection and writing
// a parsed benchSet to a db table. It closes the connection, returns the number of
// benchmark tests written, and any error encountered.
//
// A new sql transaction is created and committed per Benchmark in benchSet. This way if a
// db failure occurs all data from the benchSet is not lost.
func (benchdb *BenchDB) WriteBenchSet(benchSet parse.Set) (int, error) {
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
	// Create a transaction per Benchmark and commit it.
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

	ts := time.Now().UTC()
	// Strips of leading Benchmark string in Benchmark.Name
	name := strings.TrimPrefix(strings.TrimSpace(b.Name), "Benchmark")

	_, err = tx.Exec(q,
		ts, name, b.N, b.NsPerOp, b.AllocedBytesPerOp, b.AllocsPerOp)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Run runs all of the go test benchmarks that match regexpr in the
// current directory. By default it does not run unit tests by way of setting
// test.run to XXX. It is also responsible for parsing the benchmark set and calls
// WriteBenchSet to write the benchmark data to a postgresql database. It returns
// any error encountered.
func (benchdb *BenchDB) Run() (int, error) {
	// Exec a subprocess for go test bench command and write
	// to both stdout and a byte buffer.
	cmd := exec.Command("go", "test", "-bench", benchdb.Regex,
		"-test.run", "XXX", "-benchmem")
	var out bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &out)
	err := cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("command failed: %v", err)
	}

	// Parse stdout into a parse Set.
	benchSet, err := parse.ParseSet(&out)
	if err != nil {
		return 0, fmt.Errorf("failed to parse benchmark data: %v", err)
	}

	// Writes parse set to sql database.
	n, err := benchdb.WriteBenchSet(benchSet)
	if err != nil {
		return 0, fmt.Errorf("failed to write benchSet to db: %v", err)
	}
	return n, nil
}
