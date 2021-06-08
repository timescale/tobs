build:
		go build -o bin/tobs .

unit-tests:
		go test ./pkg/...

e2e-tests:
		./tests/e2e-tests.sh

external-db-tests:
		./tests/external-db-tests.sh

upgrade-tests:
		./tests/upgrade-tests.sh

ha-tests:
		./tests/ha-tests.sh

install-tests:
		./tests/install-tests.sh

helm-tests:
		./tests/helm-tests.sh

go-vet:
		go vet ./...

go-fmt:
		gofmt -d .

go-lint:
		golangci-lint run

all: build e2e-tests external-db-tests go-vet go-fmt go-lint