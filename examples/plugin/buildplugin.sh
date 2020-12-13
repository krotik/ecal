#!/bin/sh
go build -ldflags "-s -w" -buildmode=plugin -o myfunc.so myfunc.go
