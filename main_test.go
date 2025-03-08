// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func sendMe(t *testing.T, sig os.Signal) {
	currentPID := os.Getpid()
	currentProcess, currentProcessErr := os.FindProcess(currentPID)

	if currentProcessErr != nil {
		t.Errorf("could not get process id")
		return
	}

	if signalErr := currentProcess.Signal(sig); signalErr != nil {
		t.Errorf("could not send SIGTERM")
	}
}

func startMain(t *testing.T, args ...string) (*time.Timer, chan int) {
	// exitFunc replaces os.Exit with this function that will end main and we can catch the error here
	exitFunc = func(code int) {
		if code == 0 {
			panic(code)
		} else {
			panic(code)
		}
	}

	buildInfoTag = "test"

	// mainResult will hold the result of main
	mainReturn := make(chan int, 1)
	mainStart := make(chan struct{}, 1)

	slog.Info("starting main")
	go func() {
		os.Args = args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		defer func() {
			if r := recover(); r != nil {
				if msg, msgConv := r.(int); msgConv {
					mainReturn <- msg
				}
			}
		}()

		mainStart <- struct{}{}
		main()

		// we can just come here, if main did not call anyhow the exit function
		// normal returns from main signalize no error --> 0
		mainReturn <- 0
	}()

	<-mainStart

	slog.Info("setting exit timeout")
	afterTimer := time.AfterFunc(2*time.Second, func() {
		sendMe(t, syscall.SIGTERM)
	})

	return afterTimer, mainReturn
}

func finalizeMain(t *testing.T, afterTimer *time.Timer, result chan int) int {
	slog.Info("stoping exit timer")

	if afterTimer.Stop() {
		sendMe(t, syscall.SIGTERM)
	}

	return <-result
}

func TestSonicMain(t *testing.T) {

	afterTimer, mainReturn := startMain(t,
		"sonicweb",
		"-root", "./testroot",
		"-address", "localhost",
		"-iaddress", "localhost",
	)

	couldRequest := false

	for i := 0; i < 10 && !couldRequest; i++ {
		res, err := http.Get("http://localhost:8080/")

		if err != nil {
			runtime.Gosched()
			fmt.Printf("received error: %v\n", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		couldRequest = true

		if res.StatusCode != 200 {
			t.Errorf("Status code is not 200: %d", res.StatusCode)
		} else {
			slog.Info("request result", slog.Int("statusCode", res.StatusCode))
		}
	}

	if !couldRequest {
		t.Errorf("could not send any request")
	}

	result := finalizeMain(t, afterTimer, mainReturn)

	slog.Info("main returned", slog.Int("result", result))
}

func TestSonicMainVersion(t *testing.T) {
	afterTimer, mainReturn := startMain(t,
		"sonicweb", "-version",
	)

	runtime.Gosched()

	afterTimer.Stop()
	result := <-mainReturn

	if result != 0 {
		t.Errorf("expected successful return, but got %v", result)
	}
}

func TestSonicMainInvalidRoot(t *testing.T) {
	afterTimer, mainReturn := startMain(t,
		"sonicweb", "-root", "/noentry",
	)

	runtime.Gosched()

	afterTimer.Stop()
	result := <-mainReturn

	if result != 1 {
		t.Errorf("expected failure return, but got %v", result)
	}
}

func BenchmarkHandler(b *testing.B) {
	server := httptest.NewServer(
		generateFileHandler(
			false,
			false,
			"/",
			"testroot/"))

	client := &http.Client{}

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(server.URL)

		if err != nil {
			b.Fatalf("Failed to make GET request: %v", err)
		}

		_, err = io.Copy(io.Discard, resp.Body)

		if err != nil {
			b.Fatalf("Failed to read response body: %v", err)
		}

		_ = resp.Body.Close()
	}
}
