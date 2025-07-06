#!/bin/bash
# Convenience script for quick startup
minikube start # Start the minikube cluster first
docker compose up --build -d # Build the image and run it
docker exec -it shade_server go run main.go migrate up # Start the migrations