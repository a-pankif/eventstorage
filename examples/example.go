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

	events := eventStorage.ReadEvents(15, 1)
	//
	// for _, event := range events {
	// 	fmt.Println((event))
	// }
	// for _, event := range events {
	// 	fmt.Println(event)
	// }
	for _, event := range events {
		fmt.Println(string(event))
	}

	// data, err := eventStorage.Read(0, 99, 0)
	// fmt.Println(err)
	// decoded, _ := eventStorage.Decode(data)
	// fmt.Println(string(decoded))
}
