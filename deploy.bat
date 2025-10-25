@echo off
if "%1"=="" (
    set ADB_PORT=5555
) else (
    set ADB_PORT=%1
)

if "%2"=="" (
    set ARCH=x86
) else (
    set ARCH=%2
)

echo Deploying for Linux %ARCH%...
adb connect 127.0.0.1:%ADB_PORT%
echo Stopping httpbot on device (if running)...
adb -s 127.0.0.1:%ADB_PORT% shell "p=`pidof httpbot` && [ -n \"$p\" ] && kill -9 $p || true"
echo Pushing new binary and start script...
adb -s 127.0.0.1:%ADB_PORT% push dist\httpbot-linux-%ARCH% /data/local/tmp/httpbot
adb -s 127.0.0.1:%ADB_PORT% shell chmod +x /data/local/tmp/httpbot
adb -s 127.0.0.1:%ADB_PORT% push start-httpbot.sh /data/local/tmp/start-httpbot.sh
adb -s 127.0.0.1:%ADB_PORT% shell chmod +x /data/local/tmp/start-httpbot.sh

echo Deployment complete.
