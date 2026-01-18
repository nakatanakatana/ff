# ff

![publish-docker-image](https://github.com/nakatanakatana/ff/actions/workflows/publish-docker-image.yaml/badge.svg)
![CI](https://github.com/nakatanakatana/ff/actions/workflows/ci.yaml/badge.svg)
![Coverage](https://github.com/nakatanakatana/octocov-central/blob/main/badges/nakatanakatana/ff/coverage.svg?raw=true)
![Code to Test Ratio](https://github.com/nakatanakatana/octocov-central/blob/main/badges/nakatanakatana/ff/ratio.svg?raw=true)
![Test Execution Time](https://github.com/nakatanakatana/octocov-central/blob/main/badges/nakatanakatana/ff/time.svg?raw=true)

## query

see [test file](./filter_test.go)

## Development

### Build
```sh
go build -o dist/ ./cmd/...
```

### Lint
```sh
go tool golangci-lint run
```

### Test
```sh
go test ./... -v -coverprofile=coverage.txt -covermode=atomic
```
