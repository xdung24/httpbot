#!/bin/bash

mkdir -p dist

echo "Building for Linux x86..."
CGO_ENABLED=1 GOOS=linux GOARCH=386 CGO_CFLAGS="-Wno-error -include signal.h -Wno-unused-variable" go build -a -ldflags '-extldflags "-static"' -o dist/httpbot-linux-x86

echo "Building for Linux x64..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CGO_CFLAGS="-Wno-error -include signal.h -Wno-unused-variable" go build -a -ldflags '-extldflags "-static"' -o dist/httpbot-linux-x64

echo "Build complete."

echo "Copying dist to EvonyBot..."
mkdir -p ../EvonyBot/tools/httpbot/
cp -r dist/httpbot* ../EvonyBot/tools/httpbot/
echo "Done."