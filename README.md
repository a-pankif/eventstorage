## Install

```
go get -u github.com/pankif/eventstorage
```

## Tests
- `go tests`
- Coverage percent `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- Coverage map `go test ./ -coverprofile c.out && go tool cover -html=c.out`

## Benchmarks
- `go test -bench=. --benchmem`