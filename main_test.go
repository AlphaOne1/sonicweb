// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateCertAndKey() (string, string) {
	// generate private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	// certificate information
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Example Organization"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // validity: 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// generate self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}

	// save certificate
	certFile, err := os.CreateTemp("", "cert")
	if err != nil {
		panic(err)
	}
	defer func() { _ = certFile.Close() }()

	_ = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// save private key
	keyFile, err := os.CreateTemp("", "key")
	if err != nil {
		panic(err)
	}
	defer func() { _ = keyFile.Close() }()

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		panic(err)
	}

	_ = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	return certFile.Name(), keyFile.Name()
}

type testHelper interface {
	Helper()
	Errorf(format string, a ...any)
}

func sendMe(helper testHelper, sig os.Signal) {
	helper.Helper()

	currentPID := os.Getpid()
	currentProcess, currentProcessErr := os.FindProcess(currentPID)

	if currentProcessErr != nil {
		helper.Errorf("could not get process id")
		return
	}

	switch runtime.GOOS {
	case "windows":
		helper.Errorf("Go does not yet support anything else than KILL on Windows")
		sig = os.Kill
	default:
	}

	if signalErr := currentProcess.Signal(sig); signalErr != nil {
		helper.Errorf("could not send signal %s: %v", sig.String(), signalErr)
	}
}

func startMain(helper testHelper, args ...string) (*time.Timer, chan int) {
	helper.Helper()
	// exitFunc replaces os.Exit with this function that will end main, and we can catch the error here
	exitFunc = func(code int) {
		panic(code)
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

		// we can just come here if main did not call anyhow the exit function
		// normal returns from main signalize no error --> 0
		mainReturn <- 0
	}()

	<-mainStart

	slog.Info("setting exit timeout")
	afterTimer := time.AfterFunc(2*time.Second, func() {
		sendMe(helper, syscall.SIGTERM)
	})

	return afterTimer, mainReturn
}

func finalizeMain(h testHelper, afterTimer *time.Timer, result chan int) int {
	h.Helper()
	slog.Info("stoping exit timer")

	if afterTimer.Stop() {
		sendMe(h, syscall.SIGTERM)
	}

	return <-result
}

func TestSonicMain(t *testing.T) {
	afterTimer, mainReturn := startMain(t,
		"sonicweb",
		"-root", "./testroot",
		"-header", "X-Test-Header: testHeaderContent",
		"-header", "X-Empty",
		"-headerfile", "testroot/testHeaders.conf",
		"-tryfile", "$uri",
		"-tryfile", "/index.html",
		"-address", "localhost",
		"-iaddress", "localhost",
	)

	couldRequest := false

	for i := 0; i < 10 && !couldRequest; i++ {
		req, _ := http.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"http://localhost:8080/index.html",
			nil)
		res, err := http.DefaultClient.Do(req)

		if err != nil {
			runtime.Gosched()
			slog.Info("received error", slog.String("error", err.Error()))
			time.Sleep(500 * time.Millisecond)

			continue
		}

		defer func() {
			if err := res.Body.Close(); err != nil {
				t.Errorf("could not close response: %v", err)
			}
		}()

		couldRequest = true

		assert.Equal(t, http.StatusOK, res.StatusCode, "status code should be 200")
		assert.Equal(t,
			"testHeaderContent",
			res.Header.Get("X-Test-Header"),
			"header should contain X-Test-Header with testHeaderContent")
		assert.Contains(t,
			res.Header,
			"X-Empty",
			"X-Empty header not found")
		assert.Equal(t,
			"line0 line1",
			res.Header.Get("X-File-Test-0"),
			"header should contain X-File-Test-0")
		assert.Equal(t,
			"line2",
			res.Header.Get("X-File-Test-1"),
			"header should contain X-File-Test-1")
	}

	assert.True(t, couldRequest, "could not send any request")

	result := finalizeMain(t, afterTimer, mainReturn)

	slog.Info("main returned", slog.Int("result", result))
}

func TestSonicMainTLS(t *testing.T) {
	certFile, keyFile := generateCertAndKey()

	afterTimer, mainReturn := startMain(t,
		"sonicweb",
		"-root", "./testroot",
		"-tlscert", certFile,
		"-tlskey", keyFile,
		"-header", "X-Test-Header: testHeaderContent",
		"-header", "X-Empty",
		"-headerfile", "testroot/testHeaders.conf",
		"-tryfile", "$uri",
		"-tryfile", "/index.html",
		"-address", "localhost",
		"-iaddress", "localhost",
	)

	couldRequest := false

	for i := 0; i < 10 && !couldRequest; i++ {
		client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}

		req, _ := http.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"https://localhost:8080/index.html",
			nil)
		res, err := client.Do(req)

		if err != nil {
			runtime.Gosched()
			fmt.Printf("received error: %v\n", err)
			time.Sleep(500 * time.Millisecond)

			continue
		}

		defer func() {
			if err := res.Body.Close(); err != nil {
				t.Errorf("could not close response: %v", err)
			}
		}()

		couldRequest = true

		assert.Equal(t, http.StatusOK, res.StatusCode, "status code should be 200")
		assert.Equal(t,
			"testHeaderContent",
			res.Header.Get("X-Test-Header"),
			"header should contain X-Test-Header with testHeaderContent")
		assert.Contains(t,
			res.Header,
			"X-Empty",
			"X-Empty header not found")
		assert.Equal(t,
			"line0 line1",
			res.Header.Get("X-File-Test-0"),
			"header should contain X-File-Test-0")
		assert.Equal(t,
			"line2",
			res.Header.Get("X-File-Test-1"),
			"header should contain X-File-Test-1")
	}

	assert.True(t, couldRequest, "could not send any request")

	result := finalizeMain(t, afterTimer, mainReturn)

	slog.Info("main returned", slog.Int("result", result))

	_ = os.Remove(certFile)
	_ = os.Remove(keyFile)
}

func TestSonicMainVersion(t *testing.T) {
	afterTimer, mainReturn := startMain(t,
		"sonicweb", "-version",
	)

	runtime.Gosched()

	result := finalizeMain(t, afterTimer, mainReturn)

	assert.Equal(t, 0, result, "expected successful return")
}

func TestSonicMainInvalidRoot(t *testing.T) {
	afterTimer, mainReturn := startMain(t,
		"sonicweb",
		"-root", "/noentry",
		"-address", "localhost",
		"-iaddress", "localhost",
	)

	runtime.Gosched()

	result := finalizeMain(t, afterTimer, mainReturn)

	assert.Equal(t, 1, result, "main should exit with 1")
}

func TestSonicMainInvalidHeaderFile(t *testing.T) {
	afterTimer, mainReturn := startMain(t,
		"sonicweb",
		"-root", "testroot/",
		"-headerfile", "/noexist",
		"-address", "localhost",
		"-iaddress", "localhost",
	)

	runtime.Gosched()

	result := finalizeMain(t, afterTimer, mainReturn)

	assert.Equal(t, 1, result, "main should exit with 1")
}

func TestSonicMainInvalidWAFFile(t *testing.T) {
	afterTimer, mainReturn := startMain(t,
		"sonicweb",
		"-root", "testroot/",
		"-wafcfg", "/noexist",
		"-address", "localhost",
		"-iaddress", "localhost",
	)

	runtime.Gosched()

	result := finalizeMain(t, afterTimer, mainReturn)

	assert.Equal(t, 1, result, "main should exit with 1")
}

func BenchmarkHandler(b *testing.B) {
	fileHandler, fileHandlerErr := generateFileHandler(
		false,
		false,
		"/",
		"testroot/",
		nil,
		nil,
		nil)

	if fileHandlerErr != nil {
		b.Fatalf("could not generate file handlers: %v", fileHandlerErr)
	}

	server := httptest.NewServer(fileHandler)

	defer server.Close()

	client := &http.Client{}

	for b.Loop() {
		req, _ := http.NewRequestWithContext(
			b.Context(),
			http.MethodGet,
			server.URL,
			nil)
		resp, err := client.Do(req)

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

func sonicMainHandlerTest(t *testing.T, uri string, method string, header string, headerValue string) {
	t.Helper()

	fileHandler, fileHandlerErr := generateFileHandler(
		false,
		false,
		"/",
		"testroot/",
		nil,
		nil,
		nil)

	if fileHandlerErr != nil {
		t.Fatalf("could not generate file handlers: %v", fileHandlerErr)
	}

	// Basic input constraints to keep fuzzing focused and avoid trivial rejections
	if len(uri) == 0 || len(uri) > 1024 {
		t.Skip()
	}

	// Ensure uri is a path; strip spaces and reject NUL
	uri = strings.TrimSpace(uri)

	if strings.ContainsRune(uri, '\x00') {
		t.Skip()
	}

	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	urlParts := strings.Split(uri, "/")

	for i := range urlParts {
		urlParts[i] = url.PathEscape(urlParts[i])
	}

	uri = strings.Join(urlParts, "/")

	if _, err := url.Parse(uri); err != nil {
		t.Skip()
	}

	// Only fuzz a small, meaningful method set to reduce noise
	switch strings.ToUpper(method) {
	case "GET", "HEAD":
		method = strings.ToUpper(method)
	default:
		t.Skip()
	}

	// Validate header name; only set it if it looks sane
	validHeader := func(s string) bool {
		if s == "" || len(s) > 128 {
			return false
		}

		for i := range len(s) {
			c := s[i]
			isAlphaNum :=
				(c >= 'a' && c <= 'z') ||
					(c >= 'A' && c <= 'Z') ||
					(c >= '0' && c <= '9')

			if !isAlphaNum && c != '-' {
				return false
			}
		}

		return true
	}
	if len(headerValue) > 1024 {
		// Avoid huge header values that only stress http parsing without adding coverage
		t.Skip()
	}

	rec := httptest.NewRecorder()

	req := httptest.NewRequestWithContext(
		t.Context(),
		method,
		"http://localhost:8080"+uri,
		nil)

	if validHeader(header) {
		req.Header.Set(header, headerValue)
	}

	fileHandler.ServeHTTP(rec, req)

	if rec.Result().StatusCode >= http.StatusInternalServerError {
		t.Errorf("received status code %s (%d) for %s %s",
			http.StatusText(rec.Result().StatusCode),
			rec.Result().StatusCode,
			method,
			uri)
	}

	if err := rec.Result().Body.Close(); err != nil {
		t.Errorf("could not close response: %v", err)
	}
}

func FuzzSonicMain(f *testing.F) {
	f.Add("index.html", "GET", "X-Fuzz", "value")
	f.Add("index.html", "HEAD", "X-Empty", "")
	f.Add("%", "GET", "0", "0")
	f.Add("\xd8", "HEAD", "0", "")
	f.Add("0 0", "GET", "0", "0")
	f.Add(strings.Repeat("0", 256), "HEAD", "0", "0")

	f.Fuzz(sonicMainHandlerTest)
}

func TestSonicMainHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		uri         string
		method      string
		header      string
		headerValue string
	}{
		{uri: "%", method: "GET", header: "0", headerValue: "0"},
		{uri: "\xd8", method: "HEAD", header: "0", headerValue: ""},
		{uri: "0 0", method: "GET", header: "0", headerValue: "0"},
		{uri: strings.Repeat("0", 256), method: "HEAD", header: "0", headerValue: "0"},
	}

	for testIndex, test := range tests {
		t.Run(fmt.Sprintf("TestSonicMainHandler-%d", testIndex), func(t *testing.T) {
			t.Parallel()

			sonicMainHandlerTest(t,
				test.uri,
				test.method,
				test.header,
				test.headerValue,
			)
		})
	}
}
