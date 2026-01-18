#!/bin/zsh

rm -rf ./tests/gen

go run ../main.go gen -s config/goradd_schema.json -o ./tests/gen/goradd
go run ../main.go gen -s config/goraddunit_schema.json -o ./tests/gen/goradd_unit
