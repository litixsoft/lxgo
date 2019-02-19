#!/usr/bin/env bash

## Install dev dependencies
go get -v -u github.com/Clever/gitsem
go get -v -u github.com/go-task/task/cmd/task
go get -v -u github.com/golang/mock/mockgen
go mod tidy
go mod download