build:
		go build -o bin/tobs .

test:
	go test -v ./tests/ --timeout 30m

go-vet:
		go vet ./...

go-fmt:
		gofmt -d .

go-lint:
		golangci-lint run

all: build go-vet go-fmt go-lint