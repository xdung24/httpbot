@echo off
if "%1"=="" (
    set ADB_PORT=5555
) else (
    set ADB_PORT=%1
)
adb connect 127.0.0.1:%ADB_PORT%
adb -s 127.0.0.1:%ADB_PORT% shell "/data/local/tmp/start-httpbot.sh"