# Eventstorage

![build](https://github.com/pankif/eventstorage/actions/workflows/go.yml/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/pankif/eventstorage)](https://goreportcard.com/report/github.com/pankif/eventstorage)
[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/pankif/eventstorage/blob/main/LICENSE)

This is an event logger with high-speed recording and event reading capability. Supports log rotation.

## Benchmarks

```console
cpu: Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz  
BenchmarkWriteChar-16                        63323869       19.42 ns/op         4  B/op         0 allocs/op
BenchmarkEventStorage_ReadChar-16              323152        3665 ns/op         24 B/op         1 allocs/op
BenchmarkEventStorage_CharReadTo-16            331111        3637 ns/op          8 B/op         1 allocs/op
BenchmarkEventStorage_CharReadToOffset10000-16  43273       27300 ns/op          8 B/op         1 allocs/op
````

## Installation
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
- Coverage percent `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- Coverage map `go test ./ -coverprofile c.out && go tool cover -html=c.out`
- `go test -bench=. --benchmem`

```console
github.com/pankif/eventstorage  0.162s  coverage: 95.0% of statements
````