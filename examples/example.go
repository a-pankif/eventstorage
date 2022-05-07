package main

import (
	"fmt"
	"github.com/pankif/binarylog"
	"os"
)

func main() {
	binlogFile, _ := os.OpenFile("binlog.1", os.O_CREATE|os.O_APPEND, 0644)
	binlog := binarylog.New(binlogFile, os.Stderr, os.Stdout)
	binlog.SetAutoFlushCount(1)
	defer func() {
		_ = binlog.CloseLogFile()
	}()

	// binlog.Log([]byte("its binlog row "))

	data, _ := binlog.Read(0, 13)
	fmt.Println(string(data))
	data, _ = binlog.Read(0, 10)
	fmt.Println(string(data))
}
