#!/usr/bin/env bash

docker build . -t mitchya1/sanepa:v1-k8s1.15.7-$(git rev-parse --verify HEAD | cut -c 1-7) -t mitchya1/sanepa:v1-k8s1.15.7-latest && \
docker push mitchya1/sanepa:v1-k8s1.15.7-$(git rev-parse --verify HEAD | cut -c 1-7) && \
docker push mitchya1/sanepa:v1-k8s1.15.7-latest && \
rm sanepa