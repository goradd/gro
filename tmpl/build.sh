#!/bin/zsh

# Builds the templates

rm -rf ./template/*
got -o ./template -i ./src/*.got

rm -rf ../_test/gen
