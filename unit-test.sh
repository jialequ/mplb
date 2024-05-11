#!/bin/sh

cd core
go test ./... -coverprofile=cover.out && go tool cover -func=cover.out