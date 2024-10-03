// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestSonicMain(t *testing.T) {
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
