package main

import (
	"encoding/hex"
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

	binlog.Log([]byte("its binlog row "))

	data, _ := binlog.Read(0, 99, 0)
	decoded := binlog.Decode(data)

	fmt.Println(string(data))
	fmt.Println(decoded)
	fmt.Println(string(decoded))
}

func interest() {
	g, _ := hex.DecodeString("1") // 67 in HEX is 'g' char, 6 or 7 (or some wrong symbol) decode from hex return zero length result
	fmt.Println(string(g))
	fmt.Println(len(g))
	fmt.Println(len(string(g)))
}
