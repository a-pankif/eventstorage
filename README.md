## Install

```
go get -u github.com/pankif/eventstorage
```

## Tests coverage
- Coverage percent `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- Coverage map `go test ./ -coverprofile c.out && go tool cover -html=c.out`