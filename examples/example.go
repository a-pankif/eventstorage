package main

import (
	"fmt"
	"github.com/pankif/binarylog"
	"os"
)

func main() {
	// fmt.Println(os.TempDir())
	eventStorage, _ := binarylog.New("./", os.Stderr)
	eventStorage.SetAutoFlushCount(1)
	eventStorage.SetLogFileMaxSize(100 * binarylog.KB)

	defer func() {
		eventStorage.Shutdown()
	}()

	// written, _ := eventStorage.Log([]byte("its eventStorage row kek! "))
	//
	// fmt.Println(written)

	events := eventStorage.ReadEvents(11, 0)
	for _, event := range events {
		fmt.Println(string(event))
	}

}
