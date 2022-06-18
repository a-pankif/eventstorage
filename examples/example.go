package main

import (
	"fmt"
	"github.com/pankif/eventstorage"
	"strconv"
	"time"
)

func main() {
	storage, err := eventstorage.New("./")

	if err != nil {
		fmt.Println(err)
		return
	}

	defer storage.Shutdown()
	storage.SetAutoFlushCount(1)
	_, err = storage.Write([]byte("some event to write " + strconv.Itoa(int(time.Now().UnixMilli()))))

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(storage.Read(1, 0))
}

func fillManyFilesAndRead() {
	storage, _ := eventstorage.New("./")
	defer storage.Shutdown()

	storage.SetWriteFileMaxSize(10 * eventstorage.MB)
	_ = storage.SetAutoFlushTime(60 * time.Millisecond)

	for i := 0; i < 1100000; i++ {
		_, _ = storage.Write([]byte("event #" + strconv.Itoa(i)))
		fmt.Println(i)
	}

	time.Sleep(time.Second)

	events := storage.Read(10, 1000001)
	for _, event := range events {
		fmt.Println(event)
	}
}
