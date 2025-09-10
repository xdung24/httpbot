uinput-go

HTTP bot server for Android device automation via touch, swipe, key events, and screen capture.

Notes:
- Requires Linux or WSL to compile
- Runs on Android devices with input access

## Build

To build the project, install golang:
```sh
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
# Add to PATH to .profile or .bashrc
export PATH=$PATH:/usr/local/go/bin
```

To build this app, linux environment is required because linux head import

```sh
sudo apt-get update && sudo apt-get install -y build-essential linux-headers-generic gcc-multilib libc6-dev-i386
```

## API Endpoints

The server runs on port 8080 by default (configurable via `PORT` environment variable) and provides the following endpoints:

### GET /cap
Captures a screenshot of the device screen.

**Method:** GET  
**Response:** Binary BMP image data  
**Content-Type:** image/bmp  

**Example:**
```bash
curl -X GET http://localhost:8080/cap > screenshot.bmp
```

### POST /tap
Performs tap gestures on the screen.

**Method:** POST  
**Body:** `<x> <y> [amount] [delay]`

**Parameters:**
- `x`, `y`: Screen coordinates (required)
- `amount`: Number of taps (optional, default: 1)
- `delay`: Delay between taps in milliseconds (optional, default: 0)

**Examples:**
```bash
# Single tap at coordinates (100, 200)
curl -X POST http://localhost:8080/tap -d "100 200"

# Tap 5 times with 1000ms delay between taps
curl -X POST http://localhost:8080/tap -d "100 200 5 1000"
```

**Response:**
- Success: `OK|tapped <x> <y> <amount> <delay>`
- Error: `ERROR|<error message>`

### POST /swipe
Performs swipe gestures on the screen.

**Method:** POST  
**Body:** `<x1> <y1> <x2> <y2> <duration> <amount> <delay>`

**Parameters:**
- `x1`, `y1`: Start coordinates (required)
- `x2`, `y2`: End coordinates (required)
- `duration`: Swipe duration in milliseconds (required)
- `amount`: Number of swipes (required)
- `delay`: Delay between swipes in milliseconds (required)

**Example:**
```bash
# Swipe from (100, 200) to (300, 400) over 500ms, once, no delay
curl -X POST http://localhost:8080/swipe -d "100 200 300 400 500 1 0"
```

**Response:**
- Success: `OK|swiped <x1> <y1> <x2> <y2> <duration> <amount> <delay>`
- Error: `ERROR|<error message>`

### POST /key
Sends key events to the device.

**Method:** POST  
**Body:** `<keycode> [amount] [delay]`

**Parameters:**
- `keycode`: Android keycode (numeric or KEYCODE_* name) (required)
- `amount`: Number of key presses (optional, default: 1)
- `delay`: Delay between key presses in milliseconds (optional, default: 0)

**Examples:**
```bash
# Press back button once
curl -X POST http://localhost:8080/key -d "4"

# Press volume up 3 times with 500ms delay
curl -X POST http://localhost:8080/key -d "KEYCODE_VOLUME_UP 3 500"
```

**Response:**
- Success: `OK|key <keycode> x<amount> delay=<delay>`
- Error: `ERROR|<error message>`

### POST /text
Sends text input to the device.

**Method:** POST  
**Body:** Text string to send

**Example:**
```bash
# Send text input
curl -X POST http://localhost:8080/text -d "Hello World"
```

**Response:**
- Success: `OK|text sent len=<length>`
- Error: `ERROR|<error message>`

## Configuration

Environment variables:
- `PORT`: Server port (default: 8080)
- `WORKERS`: Number of worker threads (default: 4)
- `QUEUE_SIZE`: Action queue size (default: 128)

## Error Responses

All endpoints may return these error responses:
- `ERROR|server busy` (503): Too many concurrent requests
- `ERROR|queue full` (503): Action queue is full
- `ERROR|request body too large` (413): Request body exceeds 1KB limit

## Performance Notes

- The server uses a worker pool to process requests asynchronously
- Tap and key requests have a 150ms timeout for immediate response
- Swipe requests have a 200ms timeout
- If workers don't respond within timeout, requests return "enqueued" status (202)
- Maximum 4 concurrent requests by default (configurable)