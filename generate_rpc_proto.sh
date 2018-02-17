#!/usr/bin/env bash

#protoc -I/usr/local/include -I. \
#  -I$GOPATH/src \
#  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
#  --go_out=./ rpc/product.proto

protoc -I/usr/local/include \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --go_out=plugins=grpc:$GOPATH/src \
  $GOPATH/src/github.com/shuoyang2016/mywish/rpc/*.proto

protoc -I/usr/local/include \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --grpc-gateway_out=logtostderr=true:$GOPATH/src \
  $GOPATH/src/github.com/shuoyang2016/mywish/rpc/mywish_service.proto
