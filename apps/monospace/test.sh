#!/usr/bin/env bash

set -e
cd $(dirname "$0")

echo "" > coverage/coverage.out

go test -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
go tool cover -html=coverage/coverage.out -o coverage/coverage.html
go tool cover -func=coverage/coverage.out -o coverage/coverage.txt
