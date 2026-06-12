#!/usr/bin/env bash

if [ $# -ne 1 ]; then
    echo "Error: Missing required argument."
    echo "Usage: $0 <docker image name>"
    exit 1
fi

image_name="$1"
docker run -itd --rm -v $(pwd):/workspace -w /workspace --name "${image_name}-container" $image_name
# docker run -p 8000:8000 --rm $image_name 
