#!/usr/bin/env bash
docker login -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD}
docker build . -t mitchya1/sanepa:v1-k8s$(git branch --show-current)-$(git rev-parse --verify HEAD | cut -c 1-7) -t mitchya1/sanepa:v1-k8s$(git branch --show-current)latest && \
docker push mitchya1/sanepa:v1-k8s$(git branch --show-current)-$(git rev-parse --verify HEAD | cut -c 1-7) && \
docker push mitchya1/sanepa:v1-k8s$(git branch --show-current)-latest && \
rm -f sanepa