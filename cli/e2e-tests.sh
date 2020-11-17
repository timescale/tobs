#!/bin/sh

kind create cluster
go test -v ./tests/ --timeout 30m
kind delete cluster