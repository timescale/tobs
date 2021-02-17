build:
		go build -o bin/tobs .

e2e-tests:
		./tests/e2e-tests.sh

external-db-tests:
		./tests/external-db-tests.sh

go-vet:
		go vet ./...

go-fmt:
		gofmt -d .

go-lint:
		golangci-lint run

all: build e2e-tests external-db-tests go-vet go-fmt go-lint