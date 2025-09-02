@echo off
if "%1"=="" (
    set ADB_PORT=5555
) else (
    set ADB_PORT=%1
)
adb connect 127.0.0.1:%ADB_PORT%
set /a LOCAL_PORT = %ADB_PORT% + 1
adb -s 127.0.0.1:%ADB_PORT% forward tcp:%LOCAL_PORT% tcp:8080