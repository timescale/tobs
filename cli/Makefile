build:
		go build -o bin/tobs .

test:
		go test -v ./tests/

go-vet:
		go vet ./...

go-fmt:
		gofmt -d .

go-lint:
		golangci-lint run

all: build test go-vet go-fmt go-lint