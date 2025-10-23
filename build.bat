@echo off

mkdir dist 2>nul

echo Building for Linux x86...
set CGO_ENABLED=1
set GOOS=linux
set GOARCH=386
set CGO_CFLAGS=-Wno-error -include signal.h -Wno-unused-variable
go build -a -ldflags "-extldflags \"-static\"" -o dist/httpbot-linux-x86

echo Building for Linux x64...
set CGO_ENABLED=1
set GOOS=linux
set GOARCH=amd64
set CGO_CFLAGS=-Wno-error -include signal.h -Wno-unused-variable
go build -a -ldflags "-extldflags \"-static\"" -o dist/httpbot-linux-x64

echo Build complete.

echo Copying dist to EvonyBot...
if not exist ..\EvonyBot\tools\httpbot\ mkdir ..\EvonyBot\tools\httpbot\
xcopy dist\* ..\EvonyBot\tools\httpbot\ /E /I /Y
echo Done.