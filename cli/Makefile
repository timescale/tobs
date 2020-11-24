build:
		go build -o bin/tobs .

test:
		./e2e-tests.sh

go-vet:
		go vet ./...

go-fmt:
		gofmt -d .

go-lint:
		golangci-lint run

all: build test go-vet go-fmt go-lint