// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bufio"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/AlphaOne1/midgard/handler/add_header"
	"github.com/AlphaOne1/midgard/util"

	"github.com/corazawaf/coraza/v3"
	corhttp "github.com/corazawaf/coraza/v3/http"
)

// wafMiddleware generates the web application firewall middleware.
func wafMiddleware(configs []string) func(http.Handler) http.Handler {

	wafConfig := coraza.NewWAFConfig()

	for _, config := range configs {
		slog.Info("adding waf configuration", slog.String("config", config))
		wafConfig = wafConfig.WithDirectivesFromFile(config)
	}

	// First we initialize our waf and our seclang parser
	waf, wafErr := coraza.NewWAF(wafConfig)

	// Now we parse our rules
	if wafErr != nil {
		slog.Error("could not initialize waf", slog.String("error", wafErr.Error()))
		exitFunc(1)
	}

	return func(next http.Handler) http.Handler {
		return corhttp.WrapHandler(waf, next)
	}
}

// headerParamToHeaders takes additional headers in the form of curl, e.g. "Content-Type: application/json",
// and generates key-value pairs of them.
func headerParamToHeaders(param []string) [][2]string {
	headers := make([][2]string, 0, len(param))

	for _, p := range param {
		slog.Info("adding header", slog.String("header", p))
		s := strings.SplitN(p, ":", 2)

		if len(s) != 2 {
			s = append(s, "")
		}

		// we cut just one "  " as this is often seen after colon after the header key
		headers = append(headers, [2]string{s[0], strings.TrimSpace(s[1])})
	}

	return headers
}

// headerFilesToHeaders reads the additional header information from the given files,
// and generates key-value pairs of them.
func headerFilesToHeaders(files []string) [][2]string {
	lines := make([]string, 0, 2*len(files))

	for _, f := range files {
		slog.Info("reading additional header file", slog.String("file", f))
		fh, openErr := os.Open(f)

		if openErr != nil {
			slog.Error("could not open header file",
				slog.String("file", f),
				slog.String("error", openErr.Error()))
			exitFunc(1)
		}

		defer func() { _ = fh.Close() }()

		scanner := bufio.NewScanner(fh)

		for scanner.Scan() {
			if len(scanner.Text()) > 0 {
				// jumping comments
				if strings.HasPrefix(scanner.Text(), "#") {
					continue
				}

				// adding multi-line header content to last read header
				if strings.HasPrefix(scanner.Text(), " ") && len(lines) > 0 {
					lines[len(lines)-1] = lines[len(lines)-1] + "\n" + strings.TrimSpace(scanner.Text())
					continue
				}

				lines = append(lines, strings.TrimSpace(scanner.Text()))
			}
		}
	}

	return headerParamToHeaders(lines)
}

// addHeaders generates the header-adding middleware. It adds the Server header and all the
// additional headers given as parameter.
func addHeaders(headers [][2]string) func(http.Handler) http.Handler {
	serverVal := ServerName

	if len(buildInfoTag) > 0 {
		serverVal = serverVal + "/" + buildInfoTag
	}

	return util.Must(add_header.New(
		add_header.WithHeaders(
			append([][2]string{{"Server", serverVal}}, headers...),
		),
	))
}
