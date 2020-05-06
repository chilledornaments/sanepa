#!/usr/bin/env bash

docker run --rm \
-v ${PWD}:/go/src/github.com/mitchya1/sanepa:Z \
-w /go/src/github.com/mitchya1/sanepa \
-e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 \
golang:1.13.9 go build -v -o sanepa ./src/v1/k8s