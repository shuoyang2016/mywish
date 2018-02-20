#!/usr/bin/env bash

#scp -r ./* shuo@192.168.29.108:~/workspace/gospace/src/github.com/shuoyang2016/mywish/
./generate_rpc_proto.sh
GOARCH=amd64 GOOS=linux go install github.com/shuoyang2016/mywish
scp $GOPATH/bin/linux_amd64/mywish shuo@192.168.29.108:~/
