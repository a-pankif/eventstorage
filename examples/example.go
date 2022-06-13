package main

import (
	"fmt"
	"github.com/pankif/binarylog"
	"strconv"
	"time"
)

func main() {
	// ctx := context.Background()
	// ctx, cancel := context.WithCancel(ctx)

	eventStorage, _ := eventstorage.New("./")
	eventStorage.SetAutoFlushTime(time.Second)
	defer eventStorage.Shutdown()

	// written, _ := eventStorage.Log([]byte("its eventStorage row kek! "))
	// fmt.Println(written)

	// fmt.Println(eventStorage.ReadEvents(5, 0))
	return

	// go func() {
	// 	reader := 0
	// 	for {
	// 		evs := eventStorage.ReadEvents(1, reader)
	// 		fmt.Println(evs)
	// 		reader++
	// 		time.Sleep(5 * time.Microsecond)
	// 	}
	// }()

	go func() {
		writer := 0
		for {
			_, _ = eventStorage.Log([]byte("its eventStorage row kek! " + strconv.Itoa(writer)))
			writer++
			fmt.Println(writer)
			// time.Sleep(time.Microsecond)
		}
	}()
	go func() {
		writer := 0
		for {
			_, _ = eventStorage.Log([]byte("its eventStorage row kek! " + strconv.Itoa(writer)))
			writer++
			fmt.Println(writer)
			// time.Sleep(time.Microsecond)
		}
	}()
	go func() {
		writer := 0
		for {
			_, _ = eventStorage.Log([]byte("its eventStorage row kek! " + strconv.Itoa(writer)))
			writer++
			fmt.Println(writer)
			// time.Sleep(time.Microsecond)
		}
	}()
	go func() {
		writer := 0
		for {
			_, _ = eventStorage.Log([]byte("its eventStorage row kek! " + strconv.Itoa(writer)))
			writer++
			fmt.Println(writer)
			// time.Sleep(time.Microsecond)
		}
	}()

	fmt.Scanln()
	return
	events := eventStorage.ReadEvents(1, 0)

	for _, event := range events {
		fmt.Println(event)
	}

	// str := string(readBuffer[0:readCount])
	// fmt.Println(readCount, str)
	// for {
	// 	pos := strings.Index(str, "\n")
	// 	fmt.Println(pos, str[0:pos])
	// 	str = str[pos+1:]
	// 	if len(str) == 0 {
	// 		break
	// 	}
	// }

}
