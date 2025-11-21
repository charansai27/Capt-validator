#!/bin/bash

# Start mock server in background
python3 server.py

#wait for server to start
sleep 1

# Build Docker image
docker build -t caption-validator .

# Run SRT
docker run --rm -v "$(pwd)/samples:/data" caption-validator --t_start=0 --t_end=10 --coverage=50 --endpoint=http://host.docker.internal:8080  /data/examples.srt

# Run VTT
docker run --rm -v "$(pwd)/samples:/data" caption-validator --t_start=0 --t_end=10 --coverage=50 --endpoint=http://host.docker.internal:8080  /data/examples.vtt
