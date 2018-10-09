#!/usr/bin/env bash
# Version 1.0.0
# Install develop dependencies
go get -v github.com/Clever/gitsem
go get -v github.com/go-task/task/cmd/task
go get -v github.com/golang/mock/mockgen
#go get -v github.com/axw/gocov/gocov
#go get -v github.com/AlekSi/gocov-xml
#go get -v gopkg.in/matm/v1/gocov-html
#go get -v github.com/jstemmer/go-junit-report

## clean modules after installs
go mod tidy