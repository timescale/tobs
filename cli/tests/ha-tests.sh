#!/bin/sh

kind create cluster --name tobs

clean_up() {
  ec=$?
  kind delete cluster --name tobs
  exit $ec
}

trap clean_up SIGHUP SIGINT SIGTERM
go test -v ./tests/ha-tests --timeout 15m
clean_up