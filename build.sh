#!/bin/bash

docker build -t mangomm/go-bench-suite:latest --no-cache --build-arg VERSION=master .
docker push mangomm/go-bench-suite:latest
