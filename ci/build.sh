#!/bin/zsh

rm -rf ./tests/gen

go run ../cmd/gro-gen/main.go -s config/goradd_schema.json -o ./tests/gen/goradd
go run ../cmd/gro-gen/main.go -s config/goraddunit_schema.json -o ./tests/gen/goradd_unit
