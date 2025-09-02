@echo off
if "%1"=="" (
    set ADB_PORT=5555
) else (
    set ADB_PORT=%1
)
adb connect 127.0.0.1:%ADB_PORT%
echo Starting httpbot on device...
adb -s 127.0.0.1:%ADB_PORT% shell sh "/data/local/tmp/start-httpbot.sh"
echo "--- pidof output ---"
adb -s 127.0.0.1:%ADB_PORT% shell pidof httpbot || true
echo "--- ps output ---"
adb -s 127.0.0.1:%ADB_PORT% shell ps | grep httpbot || true