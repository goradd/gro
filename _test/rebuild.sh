#!/bin/zsh

rm -rf ../tmpl/template/*
got -o ../tmpl/template -i ../tmpl/src/*.got

rm -rf ./gen

go run ../cmd/gro-gen/main.go -s schema/goradd_schema.json -o ./gen/orm/goradd
go run ../cmd/gro-gen/main.go -s schema/goraddunit_schema.json -o ./gen/orm/goradd_unit
