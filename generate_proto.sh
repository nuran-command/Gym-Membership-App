#!/bin/bash

# Ensure GOPATH/bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Generate Protos
protoc --go_out=. --go-grpc_out=. proto/asset.proto
protoc --go_out=. --go-grpc_out=. proto/membership.proto
protoc --go_out=. --go-grpc_out=. proto/telemetry.proto
