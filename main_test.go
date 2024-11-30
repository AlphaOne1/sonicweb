// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestSonicMain(t *testing.T) {
	exitFunc = func(code int) {
		if code == 0 {
			return
		} else {
			os.Exit(code)
		}
	}

	go func() {
		os.Args = []string{
			"sonicmain.exe",
			"-root", "./testroot",
		}
		main()
	}()

	couldRequest := false

	for i := 0; i < 10; i++ {
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

		break
	}

	if !couldRequest {
		t.Errorf("could not send any request")
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
