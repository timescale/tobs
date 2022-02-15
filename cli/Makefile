KUBE_VERSION ?= 1.23
KIND_CONFIG ?= ./tests/kind-$(KUBE_VERSION).yaml

build:
		go build -o bin/tobs .

unit-tests:
		go test ./pkg/...

e2e-tests:
		go test -v ./tests/tobs-cli-tests --timeout 30m

external-db-tests:
		go test -v ./tests/external-db-tests --timeout 30m

upgrade-tests:
		go test -v ./tests/upgrade-tests

ha-tests:
		go test -v ./tests/ha-tests --timeout 20m

install-tests:
		go test -v ./tests/installation-tests --timeout 30m

helm-tests:
		go test -v ./tests/helm-tests

go-vet:
		go vet ./...

go-fmt:
		gofmt -s -w .

go-lint:
		golangci-lint run

start-kind:
		kind create cluster --name tobs --config $(KIND_CONFIG)

uninstall-tobs:
		./bin/tobs uninstall

delete-kind:
		kind delete cluster --name tobs

all: build start-kind e2e-tests external-db-tests go-vet go-fmt go-lint
