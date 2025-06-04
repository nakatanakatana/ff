# ff

![publish-docker-image](https://github.com/nakatanakatana/ff/actions/workflows/publish-docker-image.yaml/badge.svg)
![CI](https://github.com/nakatanakatana/ff/actions/workflows/ci.yaml/badge.svg)
[![codecov](https://codecov.io/gh/nakatanakatana/ff/branch/main/graph/badge.svg?token=RE9U2B89AP)](https://codecov.io/gh/nakatanakatana/ff)

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
