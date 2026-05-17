#!/bin/bash

# Ensure GOPATH/bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Generate Protos
cd proto
protoc --go_out=. --go-grpc_out=. asset.proto
protoc --go_out=. --go-grpc_out=. membership.proto
protoc --go_out=. --go-grpc_out=. telemetry.proto
cd ..
