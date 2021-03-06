package benchdb

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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
	// Run executes go test bench for benchmarks matching a regex defined in
	// BenchDBConfigthe current directory. By default it does not run unit tests
	// by way of setting test.run to XXX in the call to go test. It also parses the
	// benchSet and calls WriteSet to write the benchmark data to a database. It
	// returns any error encountered.
	Run() error

	// WriteSet is responsible for opening a postgres database connection and writing
	// a parsed benchSet to a db table. It closes the connection, returns the number of
	// benchmark tests written, and any error encountered.
	WriteSet(parse.Set) (int, error)
}

// BenchDBConfig represents configuration data for BenchDB.
type BenchDBConfig struct {
	Regex  string // regex to run tests
	ShaLen int    // number of latest git sha characters
}

// BenchPSQL represents a BenchDB that writes benchmarks to a postgres
// database.
type BenchPSQL struct {
	Config    *BenchDBConfig // configuration for go test
	Driver    string         // database driver name
	ConnStr   string         // sql connection string
	TableName string         // database table name

	dbConn *sql.DB
}

// Run runs go test benchmarks matching regex in the current directory and writes
// benchmark data to a PSQL database by calling WriteSet. It returns any error
// encountered.
func (benchdb *BenchPSQL) Run() error {
	// Exec a subprocess for go test bench and write
	// to both stdout and a byte buffer.
	cmd := exec.Command("go", "test", "-bench", benchdb.Config.Regex,
		"-test.run", "XXX", "-benchmem")
	var out bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &out)
	cmd.Stderr = io.Writer(os.Stderr)
	err := cmd.Run()
	if err != nil {
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

	batchId, err := uuid()
	if err != nil {
		return 0, fmt.Errorf("could not generate batch id: %v\n", err)
	}

	cnt := 0
	for _, b := range benchSet {
		n := len(b)
		for i := 0; i < n; i++ {
			val := b[i]
			err := benchdb.saveBenchmark(batchId, *val)
			if err != nil {
				return 0, fmt.Errorf("failed to save benchmark: %v", err)
			}
			cnt++
		}
	}
	return cnt, nil
}

func (benchdb *BenchPSQL) saveBenchmark(batchId string, b parse.Benchmark) error {
	// Create a transaction per Benchmark.
	tx, err := benchdb.dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	sha, err := latestGitSha(benchdb.Config.ShaLen)
	if err != nil {
		return err
	}

	q := fmt.Sprintf(`
        INSERT INTO %s
        (batch_id, latest_sha, datetime, name, n, ns_op, allocated_bytes_op, allocs_op)
        VALUES
        ($1, $2, $3, $4, $5, $6, $7, $8)
        `, benchdb.TableName)

	// Strips leading Benchmark string in Benchmark.Name
	name := strings.TrimPrefix(strings.TrimSpace(b.Name), "Benchmark")
	ts := time.Now().UTC()

	_, err = tx.Exec(q,
		batchId, sha, ts, name, b.N, b.NsPerOp, b.AllocedBytesPerOp, b.AllocsPerOp)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func latestGitSha(n int) (string, error) {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get latest git sha: %v\n", err)
	}
	return string(out[:n]), nil
}

func uuid() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
