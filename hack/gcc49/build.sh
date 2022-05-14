#!/bin/bash

image="ghcr.io/joyme123/gcc:4.9"

docker build -t ${image} .
docker push ${image}