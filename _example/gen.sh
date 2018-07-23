#!/usr/bin/env bash
# you can use docker instead of protoc commond
#docker run --rm -v $(pwd):$(pwd) -w $(pwd) naresti/protoc --go_out=plugins=grpc:. --grpc-gateway_out=logtostderr=false:. --rest_out=plugins=rest:.   -I. pb/*.proto

protoc -I/usr/local/include -I. \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  -I$GOPATH/src/github.com/google/protobuf/src \
  --go_out=plugins=grpc:. \
  pb/*.proto

protoc -I/usr/local/include -I. \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  -I$GOPATH/src/github.com/google/protobuf/src \
  --grpc-gateway_out=logtostderr=false:. \
  pb/*.proto

protoc -I/usr/local/include -I. \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  -I$GOPATH/src/github.com/google/protobuf/src \
  --rest_out=plugins=rest:. \
  pb/*.proto
