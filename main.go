package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

// tap command format: tap <x> <y> <amount> <delay>
// example: tap 100 200 5 1000
func handleTapCommand(cmdStr string) string {
	coords := strings.TrimPrefix(cmdStr, "tap ")
	parts := strings.Split(coords, " ")
	if len(parts) < 2 {
		return "ERROR|invalid tap command format, expected: tap <x> <y> <amount> <delay>"
	}
	x := parts[0]
	y := parts[1]

	amount := 1 // default amount is 1
	if len(parts) > 2 {
		parsedAmount, err := strconv.Atoi(parts[2])
		if err != nil || parsedAmount < 1 {
			return fmt.Sprintf("ERROR|invalid amount value: %v", err)
		}
		amount = parsedAmount
	}
	delay := 0
	if len(parts) > 3 {
		parsedDelay, err := strconv.Atoi(parts[3])
		if err != nil || parsedDelay < 0 {
			return fmt.Sprintf("ERROR|invalid delay value: %v", err)
		}
		delay = parsedDelay
	}
	for i := 0; i < amount; i++ {
		// execute the tap command using adb shell input tap <x> <y>
		cmd := exec.Command("/system/bin/sh", "-c", fmt.Sprintf("input tap %s %s", x, y))
		_, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("ERROR|executing command failed: %v", err)
		}
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
	return fmt.Sprintf("OK|tapped %s %s %d %d", x, y, amount, delay)
}

// swipe command format: swipe <x1> <y1> <x2> <y2> <duration>
func handleSwipeCommand(cmdStr string) string {
	coords := strings.TrimPrefix(cmdStr, "swipe ")
	parts := strings.Split(coords, " ")
	if len(parts) != 5 {
		return "ERROR|invalid swipe command format, expected: swipe <x1> <y1> <x2> <y2> <duration>"
	}
	x1 := parts[0]
	y1 := parts[1]
	x2 := parts[2]
	y2 := parts[3]
	duration := parts[4]
	// execute the swipe command using adb shell input swipe <x1> <y1> <x2> <y2> <duration>
	cmd := exec.Command("/system/bin/sh", "-c", fmt.Sprintf("input swipe %s %s %s %s %s", x1, y1, x2, y2, duration))
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("ERROR|executing command failed: %v", err)
	}
	return fmt.Sprintf("OK|swiped %s %s %s %s %s", x1, y1, x2, y2, duration)
}

func main() {
	handler := func(ctx *fasthttp.RequestCtx) {
		if string(ctx.Method()) == "GET" && string(ctx.Path()) == "/cap" {
			// GET /cap return bitmap image of screenshot
			ctx.SetContentType("image/bmp")
			cmdCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			cmd := exec.CommandContext(cmdCtx, "/data/local/tmp/ascreencap", "--stdout")
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SetBodyString("Error starting command")
				return
			}
			if err := cmd.Start(); err != nil {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SetBodyString("Error running command")
				return
			}
			// Use buffered reader for faster I/O
			reader := bufio.NewReaderSize(stdout, 256*1024) // 256KB buffer
			io.Copy(ctx, reader)
			cmd.Wait()
		} else if string(ctx.Method()) == "POST" && string(ctx.Path()) == "/swipe" {
			// POST /swipe <action>

			actionStr := string(ctx.PostBody())
			result := handleSwipeCommand(actionStr)
			ctx.SetBodyString(result)
		} else if string(ctx.Method()) == "POST" && string(ctx.Path()) == "/tap" {
			// POST /tap <action>

			actionStr := string(ctx.PostBody())
			result := handleTapCommand(actionStr)
			ctx.SetBodyString(result)
		} else {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	}

	log.Printf("Starting server on port 8080")
	log.Fatal(fasthttp.ListenAndServe(":8080", handler))
}
