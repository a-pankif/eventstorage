## Install

```
go get -u github.com/pankif/eventstorage
```

## Examples
More examples you can find into [examples](https://github.com/pankif/eventstorage/tree/main/examples).

```go
package main

import (
	"fmt"
	"time"
	"github.com/pankif/eventstorage"
)

func main()  {
	storage, err := eventstorage.New("./")
	defer storage.Shutdown()
	
	if err != nil {
		fmt.Println(err)
		return
    }

	storage.SetWriteFileMaxSize(10 * eventstorage.MB)
	storage.SetAutoFlushCount(1)
	_ = storage.SetAutoFlushTime(60 * time.Millisecond)

	_, _ = storage.Write([]byte("some data to write"))

    fmt.Println(storage.Read(1, 0)) 
}
```

## Tests
- `go tests`
- Coverage percent `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- Coverage map `go test ./ -coverprofile c.out && go tool cover -html=c.out`

```console
github.com/pankif/eventstorage  0.162s  coverage: 95.0% of statements
````

## Benchmarks
`go test -bench=. --benchmem`

```console
cpu: Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz  
BenchmarkWrite-16                               34041297             33.02 ns/op              86 B/op          0 allocs/op
BenchmarkEventStorage_Read-16                     424041              2822 ns/op              72 B/op          4 allocs/op`
BenchmarkEventStorage_ReadTo-16                   425503              2802 ns/op             139 B/op          3 allocs/op
BenchmarkEventStorage_ReadToOffset-16                181           6690154 ns/op          560102 B/op      30003 allocs/op
````