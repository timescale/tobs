#!/bin/sh

kind create cluster --name tobs

clean_up() {
  ec=$?
  kind delete cluster --name tobs
  exit $ec
}

trap clean_up SIGHUP SIGINT SIGTERM
go test -v ./tests/upgrade-tests
clean_up