#!/usr/bin/env bash

docker build . -t mitchya1/sanepa:v1-k8s1.16.8-$(git rev-parse --verify HEAD | cut -c 1-7) -t mitchya1/sanepa:v1-k8s1.16.8-latest && \
docker push mitchya1/sanepa:v1-k8s1.16.8-$(git rev-parse --verify HEAD | cut -c 1-7) && \
docker push mitchya1/sanepa:v1-k8s1.16.8-latest && \
rm -f sanepa