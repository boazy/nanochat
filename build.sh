#!/bin/sh

cd cmd/nanochat
go get
go install
cd ../..
cd cmd/nanochatd
go get
go install

