@echo off

mkdir build

echo Building for Linux x86...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=386
go build -o build\httpbot-linux-x86 main.go

echo Building for Linux ARMv7...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm
set GOARM=7
go build -o build\httpbot-linux-armv7 main.go

echo Building for Linux ARMv8...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm64
go build -o build\httpbot-linux-armv8 main.go

echo Build complete.
