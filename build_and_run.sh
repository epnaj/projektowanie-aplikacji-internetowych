#!/usr/bin/env bash

image_name="pixel-tracker-dev"

docker build -t $image_name .

if [[ "${1}" == "--run" ]]; then
    ./run-docker.sh $image_name
fi

