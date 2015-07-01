# gobenchdb
Store go test bench data in a database

[![GoDoc](https://godoc.org/github.com/yhat/gobenchdb?status.svg)](https://godoc.org/github.com/yhat/gobenchdb)

gobenchdb is a command line tool for running and storing go benchmark data in a database.
It runs the `go test -bench` command in the current working directory and parses the output
using the [parse package](https://godoc.org/golang.org/x/tools/benchmark/parse). The parsed
data is then written to a sql database of your choice. 

Writing benchmark tests in go is simple. The `go test -bench` command is great, but what we needed was a simple tool that organizes the benchmarking data it produces across multiple benchmarking test suites and as the source code changes over time.

# Installation

If you have the go tools installed on your machine, gobenchdb can be installed using `go get`.

```
go get github.com/yhat/gobenchdb
```

Direct downloads of compiled binaries are available at the [releases page](https://github.com/yhat/gobenchdb/releases).

# Basic Usage

gobenchdb supports sqlite3 and postgres as sql database backends. 

```
Usage: gobenchdb [options...]

Options:
  -conn        sql database connection string
  -table       sql table name
  -test.bench  run only those benchmarks matching the regular expression
```

# Example

You can cd to any package directory that has defined benchmark tests and run gobenchdb. Lets
run a few benchmarks from the golang crypto package and store them in a database!

```
cd $GOPATH/golang.org/src/golang.org/x/crypto/ssh
```

gobenchdb writes to stdout and a database associated with your connection string. Here is
what you get when you run the command with a connection string and a sql database table
name.

```
$ gobenchdb --conn="postgres://yhat:foopass@/benchmarks" --table="mytable"
PASS
BenchmarkEndToEnd	       100	  10195771 ns/op    102.84 MB/s    1286656 B/op	      78 allocs/op
BenchmarkMarshalKexInitMsg     200000	  8956 ns/op	    4040 B/op	   7 allocs/op
BenchmarkUnmarshalKexInitMsg   100000	  22160 ns/op       5392 B/op	   43 allocs/op
BenchmarkMarshalKexDHInitMsg   1000000	  1165 ns/op	    248 B/op	   8 allocs/op
BenchmarkUnmarshalKexDHInitMsg 2000000    658 ns/op         96 B/op	   4 allocs/op
ok  				   golang.org/x/crypto/ssh	  8.595s
```

Lets look at our database.

```
benchmarks=> select * from mytable where latest_sha = 'c57d4a7';
 id |             batch_id             | latest_sha |          datetime          |         name          |    n    |  ns_op   | allocated_bytes_op | allocs_op 
----+----------------------------------+------------+----------------------------+-----------------------+---------+----------+--------------------+-----------
 48 | e1ae21896edb38420d767cace4957efe | c57d4a7    | 2015-07-01 15:39:32.00976  | EndToEnd              |     100 | 10257787 |            1286660 |        78
 49 | e1ae21896edb38420d767cace4957efe | c57d4a7    | 2015-07-01 15:39:32.105    | MarshalKexInitMsg     |  200000 |     9099 |               4040 |         7
 50 | e1ae21896edb38420d767cace4957efe | c57d4a7    | 2015-07-01 15:39:32.189808 | UnmarshalKexInitMsg   |  100000 |    22147 |               5392 |       143
 51 | e1ae21896edb38420d767cace4957efe | c57d4a7    | 2015-07-01 15:39:32.270077 | MarshalKexDHInitMsg   | 1000000 |     1161 |                248 |         8
 52 | e1ae21896edb38420d767cace4957efe | c57d4a7    | 2015-07-01 15:39:32.350378 | UnmarshalKexDHInitMsg | 2000000 |      660 |                 96 |         4

```

Each gobenchdb run is assigned a unique batch_id and the first 7 characters of the latest git sha for HEAD. This way you can group by a batch_id and
identify separate benchmark test runs.

Thats it!