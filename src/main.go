package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/valyala/fasthttp"
)

// semaphore for concurrency limiting
var sem = make(chan struct{}, 4)

// Action represents a device action to be processed by worker pool.
type Action struct {
	Cmd   string // original command string
	Type  string // "tap" or "swipe"
	Reply chan string
}

var (
	// actions channel used by HTTP handlers to enqueue work
	actions chan *Action
)

const devicePath = "/dev/input/event5" // touch screen

// tap command format: <x> <y> <amount> <delay>
// example: 100 200 5 1000
// parseTap parses a tap command string and returns integer values or error.
// Expected formats:
// "<x> <y>" (defaults amount=1, delay=0)
// "<x> <y> <amount> <delay>"
func parseTap(cmdStr string) (x, y, amount, delay int, err error) {
	coords := strings.TrimSpace(cmdStr)
	parts := strings.Fields(coords)
	if len(parts) != 2 && len(parts) != 4 {
		return 0, 0, 0, 0, fmt.Errorf("invalid tap command format, expected: <x> <y> [amount] [delay]")
	}
	x, err = strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	y, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	amount = 1
	if len(parts) > 2 {
		amount, err = strconv.Atoi(parts[2])
		if err != nil || amount < 1 {
			return 0, 0, 0, 0, fmt.Errorf("invalid amount")
		}
	}
	delay = 0
	if len(parts) > 3 {
		delay, err = strconv.Atoi(parts[3])
		if err != nil || delay < 0 {
			return 0, 0, 0, 0, fmt.Errorf("invalid delay")
		}
	}
	return
}

func handleTapCommand(cmdStr string) string {
	x, y, amount, delay, err := parseTap(cmdStr)
	if err != nil {
		return "ERROR|" + err.Error()
	}
	// run taps with per-call timeout and no shell to reduce injection surface
	dev, err := OpenInputDevice(devicePath)
	if err != nil {
		log.Printf("open device: %v", err)
	}
	defer dev.Close()
	for i := 0; i < amount; i++ {
		// out, err := runInputCommand(3*time.Second, "tap", strconv.Itoa(x), strconv.Itoa(y))
		// if err != nil {
		// 	return fmt.Sprintf("ERROR|executing command failed: %v output:%s", err, out)
		// }
		if err := dev.SendTouch(x*60, y*33); err != nil {
			log.Printf("send touch: %v", err)
		}
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
	_ = amount
	return fmt.Sprintf("OK|tapped %d %d %d %d", x, y, amount, delay)
}

// swipe command format: <x1> <y1> <x2> <y2> <duration> <amount> <delay>
// parseSwipe parses swipe command string: x1 y1 x2 y2 duration amount delay
func parseSwipe(cmdStr string) (x1, y1, x2, y2, duration, amount, delay int, err error) {
	coords := strings.TrimSpace(cmdStr)
	parts := strings.Fields(coords)
	if len(parts) != 7 && len(parts) != 5 {
		return 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid swipe command format, expected: <x1> <y1> <x2> <y2> <duration> [amount] [delay]")
	}
	x1, err = strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	y1, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	x2, err = strconv.Atoi(parts[2])
	if err != nil {
		return
	}
	y2, err = strconv.Atoi(parts[3])
	if err != nil {
		return
	}
	duration, err = strconv.Atoi(parts[4])
	if err != nil || duration < 0 {
		return 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid duration")
	}
	amount, err = strconv.Atoi(parts[5])
	if err != nil || amount < 1 {
		return 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid amount")
	}
	delay, err = strconv.Atoi(parts[6])
	if err != nil || delay < 0 {
		return 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid delay")
	}
	return
}

func handleSwipeCommand(cmdStr string) string {
	x1, y1, x2, y2, duration, amount, delay, err := parseSwipe(cmdStr)
	if err != nil {
		return "ERROR|" + err.Error()
	}
	// execute swipe amount times using helper to run the input jar
	for i := 0; i < amount; i++ {
		out, err := runInputCommand(3*time.Second,
			"swipe", strconv.Itoa(x1), strconv.Itoa(y1), strconv.Itoa(x2), strconv.Itoa(y2), strconv.Itoa(duration),
		)
		if err != nil {
			return fmt.Sprintf("ERROR|executing command failed: %v output:%s", err, out)
		}
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
	return fmt.Sprintf("OK|swiped %d %d %d %d %d %d %d", x1, y1, x2, y2, duration, amount, delay)
}

// parseKey parses key command string: <keycode> [amount] [delay]
// keycode can be numeric or a KEYCODE_* name
func parseKey(cmdStr string) (keycode string, amount, delay int, err error) {
	coords := strings.TrimSpace(cmdStr)
	if coords == "" {
		// also support sendKey prefix
		coords = strings.TrimSpace(strings.TrimPrefix(cmdStr, "sendKey"))
	}
	parts := strings.Fields(coords)
	if len(parts) != 1 && len(parts) != 3 {
		return "", 0, 0, fmt.Errorf("invalid key command format, expected: <keycode> [amount] [delay]")
	}
	keycode = parts[0]
	amount = 1
	if len(parts) > 1 {
		amount, err = strconv.Atoi(parts[1])
		if err != nil || amount < 1 {
			return "", 0, 0, fmt.Errorf("invalid amount")
		}
	}
	delay = 0
	if len(parts) > 2 {
		delay, err = strconv.Atoi(parts[2])
		if err != nil || delay < 0 {
			return "", 0, 0, fmt.Errorf("invalid delay")
		}
	}
	return
}

func handleKeyCommand(cmdStr string) string {
	keycode, amount, delay, err := parseKey(cmdStr)
	if err != nil {
		return "ERROR|" + err.Error()
	}
	for i := 0; i < amount; i++ {
		out, err := runInputCommand(3*time.Second, "keyevent", keycode)
		if err != nil {
			return fmt.Sprintf("ERROR|executing command failed: %v output:%s", err, out)
		}
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
	return fmt.Sprintf("OK|key %s x%d delay=%d", keycode, amount, delay)
}

func handleTextCommand(cmdStr string) string {
	if cmdStr == "" {
		return "ERROR|empty text"
	}
	out, err := runInputCommand(5*time.Second, "text", cmdStr)
	if err != nil {
		return fmt.Sprintf("ERROR|executing command failed: %v output:%s", err, out)
	}
	return fmt.Sprintf("OK|text sent len=%d", len(cmdStr))
}

// runInputCommand runs the Android input command via app_process with a per-call timeout.
// It returns the combined output as a string and any execution error.
func runInputCommand(timeout time.Duration, params ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	args := append([]string{"/system/bin", "com.android.commands.input.Input"}, params...)
	cmd := exec.CommandContext(ctx, "/system/bin/app_process", args...)
	cmd.Env = append(os.Environ(), "CLASSPATH=/system/framework/input.jar")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// worker processes actions from the queue. It attempts to execute quickly
// and sends back the result on the Action.Reply channel.
func worker() {
	for a := range actions {
		var res string
		switch a.Type {
		case "tap":
			res = handleTapCommand(a.Cmd)
		case "swipe":
			res = handleSwipeCommand(a.Cmd)
		case "key":
			res = handleKeyCommand(a.Cmd)
		case "text":
			res = handleTextCommand(a.Cmd)
		default:
			res = "ERROR|unknown action"
		}
		// non-blocking send; if client isn't listening, drop result
		select {
		case a.Reply <- res:
		default:
		}
	}
}

// backgroundCapture periodically captures screenshot bytes into cachedCapture
// so /cap can return most recent image quickly.
// background capture removed; /cap will run synchronous capture when enabled.

// Pre-allocated responses for better performance
var (
	healthResponse  = []byte(`{"status":"ok"}`)
	contentTypeJSON = []byte("application/json")
)

func main() {
	// Basic logging flags
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// Config
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// config worker pool & capture
	workers := 4
	if v := os.Getenv("WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			workers = n
		}
	}
	queueSize := 128
	if v := os.Getenv("QUEUE_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			queueSize = n
		}
	}

	// create action queue and start workers
	actions = make(chan *Action, queueSize)
	for i := 0; i < workers; i++ {
		go func() {
			worker()
		}()
	}

	// create fasthttp server for graceful shutdown
	handler := func(ctx *fasthttp.RequestCtx) {
		// simple routing
		method := string(ctx.Method())
		path := string(ctx.Path())

		// limit body size for POST endpoints
		const maxBody = 1024 // 1KB
		if method == "POST" && len(ctx.PostBody()) > maxBody {
			ctx.SetStatusCode(fasthttp.StatusRequestEntityTooLarge)
			ctx.SetBodyString("ERROR|request body too large")
			return
		}

		// concurrency limiter
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		default:
			// metrics disabled: would increment errorCount
			ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
			ctx.SetBodyString("ERROR|server busy")
			return
		}

		switch {
		case method == "GET" && path == "/cap":
			ctx.SetContentType("image/bmp")
			cmdCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(cmdCtx, "/data/local/tmp/ascreencap", "--stdout")
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SetBodyString("ERROR|starting command")
				return
			}
			if err := cmd.Start(); err != nil {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SetBodyString("ERROR|running command")
				return
			}
			reader := bufio.NewReaderSize(stdout, 256*1024) // 256KB buffer
			io.Copy(ctx, reader)
			cmd.Wait()

		case method == "POST" && path == "/swipe":
			// enqueue swipe action for background worker to lower request latency
			actionStr := string(ctx.PostBody())
			act := &Action{Cmd: actionStr, Type: "swipe", Reply: make(chan string, 1)}
			select {
			case actions <- act:
				// wait briefly for worker result
				select {
				case res := <-act.Reply:
					ctx.SetBodyString(res)
				case <-time.After(200 * time.Millisecond):
					// worker didn't respond fast; return accepted
					ctx.SetStatusCode(fasthttp.StatusAccepted)
					ctx.SetBodyString("OK|swipe enqueued")
				}
			default:
				ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
				ctx.SetBodyString("ERROR|queue full")
			}

		case method == "POST" && path == "/tap":
			// enqueue tap action for background worker to lower request latency
			actionStr := string(ctx.PostBody())
			act := &Action{Cmd: actionStr, Type: "tap", Reply: make(chan string, 1)}
			select {
			case actions <- act:
				select {
				case res := <-act.Reply:
					ctx.SetBodyString(res)
				case <-time.After(150 * time.Millisecond):
					// immediate response to minimize latency
					ctx.SetStatusCode(fasthttp.StatusAccepted)
					ctx.SetBodyString("OK|tap enqueued")
				}
			default:
				ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
				ctx.SetBodyString("ERROR|queue full")
			}

		case method == "POST" && path == "/key":
			// enqueue key action
			actionStr := string(ctx.PostBody())
			act := &Action{Cmd: actionStr, Type: "key", Reply: make(chan string, 1)}
			select {
			case actions <- act:
				select {
				case res := <-act.Reply:
					ctx.SetBodyString(res)
				case <-time.After(150 * time.Millisecond):
					ctx.SetStatusCode(fasthttp.StatusAccepted)
					ctx.SetBodyString("OK|key enqueued")
				}
			default:
				ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
				ctx.SetBodyString("ERROR|queue full")
			}

		case method == "POST" && path == "/text":
			// enqueue text action
			actionStr := string(ctx.PostBody())
			act := &Action{Cmd: actionStr, Type: "text", Reply: make(chan string, 1)}
			select {
			case actions <- act:
				select {
				case res := <-act.Reply:
					ctx.SetBodyString(res)
				case <-time.After(150 * time.Millisecond):
					ctx.SetStatusCode(fasthttp.StatusAccepted)
					ctx.SetBodyString("OK|text enqueued")
				}
			default:
				ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
				ctx.SetBodyString("ERROR|queue full")
			}

		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	}

	srv := &fasthttp.Server{
		Handler:            handler,
		ReadTimeout:        5 * time.Second,  // Reduced from 15s
		WriteTimeout:       5 * time.Second,  // Reduced from 15s
		IdleTimeout:        30 * time.Second, // Reduced from 60s
		MaxConnsPerIP:      1000,
		MaxRequestsPerConn: 1000,
		TCPKeepalive:       true,
		DisableKeepalive:   false,
		ReadBufferSize:     4096,  // Smaller buffer for lower latency
		WriteBufferSize:    4096,  // Smaller buffer for lower latency
		ReduceMemoryUsage:  false, // Prioritize speed over memory
		Name:               "httpbot",
	}

	go func() {
		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(":" + port); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	// graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Printf("Shutting down server")
	srv.Shutdown()
}
