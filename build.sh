#!/bin/bash

docker build -t mangomm/go-bench-suite:latest --build-arg VERSION=master .
docker push mangomm/go-bench-suite:latest
