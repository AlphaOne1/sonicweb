// Copyright the SonicWeb contributors.
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
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateCertAndKey() (string, string) {
	// Private Key erstellen
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	// Zertifikatsinformationen
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Example Organization"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Gültigkeit: 1 Jahr
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Selbstsigniertes Zertifikat erstellen
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}

	// Zertifikat speichern
	certFile, err := os.CreateTemp("", "cert")
	if err != nil {
		panic(err)
	}
	defer func() { _ = certFile.Close() }()

	_ = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Privaten Schlüssel speichern
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

func sendMe(t *testing.T, sig os.Signal) {
	currentPID := os.Getpid()
	currentProcess, currentProcessErr := os.FindProcess(currentPID)

	if currentProcessErr != nil {
		t.Errorf("could not get process id")
		return
	}

	if signalErr := currentProcess.Signal(sig); signalErr != nil {
		t.Errorf("could not send signal")
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

	runtime.Gosched()
	time.Sleep(10 * time.Millisecond)

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
		res, err := http.Get("http://localhost:8080/index.html")

		if err != nil {
			runtime.Gosched()
			fmt.Printf("received error: %v\n", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

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
		res, err := client.Get("https://localhost:8080/index.html")

		if err != nil {
			runtime.Gosched()
			fmt.Printf("received error: %v\n", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

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
	server := httptest.NewServer(
		generateFileHandler(
			false,
			false,
			"/",
			"testroot/",
			nil,
			nil,
			nil))

	defer server.Close()

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
