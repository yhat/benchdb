package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/tools/benchmark/parse"
)

func runBenchmark(regexpr string) (parse.Set, error) {
	cmd := exec.Command("go", "test", "-bench", regexpr, "-test.run", "XXX", "-benchmem")
	var out bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &out)
	err := cmd.Run()
	if err != nil {
		log.Printf("command failed: %v", err)
		return nil, err
	}
	bench, err := parse.ParseSet(&out)
	if err != nil {
		log.Printf("failed to parse benchmark data: %v", err)
		return nil, err
	}
	return bench, nil
}

func main() {
	suite := flag.String("-test.bench", ".", "benchmark regexp")
	flag.Parse()
	b, err := runBenchmark(*suite)
	if err != nil {
		fmt.Printf("%v", err)
	}
	writer := csv.NewWriter(os.Stdout)
	for _, v := range b {
		n := len(v)
		for i := 0; i < n; i++ {
			row := strings.Split(v[i].String(), " ")
			err = writer.Write(row)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
		}
	}
	writer.Flush()
}
