#!/bin/bash

mkdir -p dist

echo "Building for Linux x86..."
CGO_ENABLED=1 GOOS=linux GOARCH=386 go build -a -ldflags '-extldflags "-static"' -o dist/httpbot-linux-x86

echo "Building for Linux x64..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o dist/httpbot-linux-x64

# echo "Building for Linux ARMv7..."
# CGO_ENABLED=1 GOOS=linux GOARCH=arm go build -a -ldflags '-extldflags "-static"' -o dist/httpbot-linux-armv7

# echo "Building for Linux ARMv8..."
# CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -a -ldflags '-extldflags "-static"' -o dist/httpbot-linux-armv8

echo "Build complete."

echo "Copying dist to EvonyBot..."
cp -r dist ../EvonyBot/httpbot