// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlphaOne1/midgard/handler/add_header"
	"github.com/AlphaOne1/midgard/util"

	"github.com/corazawaf/coraza/v3"
	corhttp "github.com/corazawaf/coraza/v3/http"
)

// wafMiddleware generates the web application firewall middleware.
func wafMiddleware(configs []string) (func(http.Handler) http.Handler, error) {
	wafConfig := coraza.NewWAFConfig()

	for _, config := range configs {
		slog.Info("adding waf configuration", slog.String("config", config))
		wafConfig = wafConfig.WithDirectivesFromFile(config)
	}

	// First, we initialize our waf and our seclang parser
	waf, wafErr := coraza.NewWAF(wafConfig)

	// Now we parse our rules
	if wafErr != nil {
		return nil, fmt.Errorf("could not initialize waf %w", wafErr)
	}

	return func(next http.Handler) http.Handler {
		return corhttp.WrapHandler(waf, next)
	}, nil
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

// headerFilesToHeaders reads the additional header information from the given files
// and generates key-value pairs of them.
func headerFilesToHeaders(files []string) ([][2]string, error) {
	var allLines []string

	for _, filePath := range files {
		slog.Info("reading additional header file", slog.String("file", filePath))

		lines, err := readHeaderFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("could not process header file %v: %w", filePath, err)
		}

		allLines = append(allLines, lines...)
	}

	return headerParamToHeaders(allLines), nil
}

// readHeaderFile opens and reads a header file, returning the parsed header lines.
func readHeaderFile(filePath string) ([]string, error) {
	fh, err := os.Open(filepath.Clean(filePath))

	if err != nil {
		return nil, fmt.Errorf("could not open file %v: %w", filePath, err)
	}

	defer func() { _ = fh.Close() }()

	return parseHeaderLines(fh)
}

// parseHeaderLines scans the reader and extracts header lines, handling comments
// and multi-line headers.
func parseHeaderLines(reader io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			continue
		}

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Handle multi-line header content
		if strings.HasPrefix(line, " ") && len(lines) > 0 {
			lines[len(lines)-1] = lines[len(lines)-1] + "\n" + strings.TrimSpace(line)
			continue
		}

		lines = append(lines, strings.TrimSpace(line))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %w", err)
	}

	return lines, nil
}

// addHeaders generates the header-adding middleware. It adds the Server header and all the
// additional headers given as parameters.
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

// addTryFiles looks if the given URI matches an existing file.
// If there is no file, a series of other files is tried instead.
func addTryFiles(tries []string, fileSystem fs.StatFS) func(http.Handler) http.Handler {
	tryFiles := make([]string, 0, len(tries))

	for _, v := range tries {
		slog.Info("registering try files", slog.String("pattern", v))

		// preventing endless loops due to file handler redirecting /index.html to /
		if strings.HasSuffix(v, "/index.html") {
			v = v[:len(v)-len("index.html")]
		}

		tryFiles = append(tryFiles, v)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expandFunc := func(s string) string {
				switch s {
				case "uri":
					return r.URL.Path
				// case "query_params":
				//		return r.URL.RawQuery
				default:
					slog.Warn("unknown variable in tryfile", slog.String("name", s))
				}

				return ""
			}

			for _, filename := range tryFiles {
				path := os.Expand(filename, expandFunc)

				if len(path) == 0 {
					// since we operate on os.Root, we can have leading /
					path = "/"
				}

				if _, statErr := fileSystem.Stat(strings.TrimPrefix(path, "/")); statErr != nil {
					if strings.HasSuffix(path, "/") {
						if _, statErr2 := fileSystem.Stat(strings.TrimPrefix(path, "/") + "index.html"); statErr2 != nil {
							// path does not exist, has / suffix, and also path/index.html does not exist
							continue
						}
					} else {
						// path does not have /-suffix, and does not exist
						continue
					}
				}

				slog.Debug("using try file",
					slog.String("pattern", filename),
					slog.String("path", path))

				r.URL.Path = path

				next.ServeHTTP(w, r)

				return
			}

			slog.Debug("no try-files matched")
			next.ServeHTTP(w, r)
		})
	}
}
