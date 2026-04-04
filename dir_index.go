package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed dir_index.html.tmpl
var directoryListingTemplate string

type FileEntry struct {
	Name       string
	Info       fs.FileInfo
	LinkTarget string
}

// cleanRequestPath cleans the URL path by trimming leading and trailing slashes.
func cleanRequestPath(urlPath string) string {
	return strings.TrimSuffix(strings.TrimPrefix(urlPath, "/"), "/")
}

// hasIndexFile checks if an index.html file exists in the given directory path.
func hasIndexFile(fsys fs.StatFS, path string) bool {
	indexPath := strings.TrimLeft(filepath.Join(path, "/index.html"), "./")

	if index, err := fsys.Stat(indexPath); err == nil && !index.IsDir() {
		return true
	}

	return false
}

// collectDirectoryEntries reads and processes directory entries, handling symlinks.
func collectDirectoryEntries(fsys fs.StatFS, path, basePath string) ([]FileEntry, error) {
	rawEntries, err := fs.ReadDir(fsys, path)
	if err != nil {
		return nil, err
	}

	entries := make([]FileEntry, 0, len(rawEntries))

	for _, rawEntry := range rawEntries {
		finfo, err := rawEntry.Info()

		if err != nil {
			continue
		}

		linkTarget := ""

		if rawEntry.Type()&fs.ModeSymlink == fs.ModeSymlink {
			lntgt, err := fs.ReadLink(fsys, rawEntry.Name())

			if err != nil {
				continue
			}

			if _, err := fsys.Stat(lntgt); err != nil {
				continue
			}

			resolvedTarget := lntgt

			if !filepath.IsAbs(lntgt) {
				resolvedTarget = filepath.Join(path, lntgt)
				resolvedTarget = filepath.Clean(resolvedTarget)
			}

			linkTarget = filepath.Join(basePath, resolvedTarget)
			linkTarget = "/" + strings.TrimPrefix(strings.ReplaceAll(linkTarget, "\\", "/"), "/")
		}

		entries = append(entries, FileEntry{
			Name:       rawEntry.Name(),
			Info:       finfo,
			LinkTarget: linkTarget,
		})
	}

	return entries, nil
}

// buildDirectoryListingParams builds the parameters for the directory listing template.
func buildDirectoryListingParams(path, basePath string, entries []FileEntry, r *http.Request) map[string]any {
	params := map[string]any{
		"DirectoryName":   filepath.Join(basePath, path),
		"DirectoryPrefix": filepath.Join(basePath, path),
		"Entries":         entries,
	}

	if path != "." {
		parentDir := basePath

		if idx := strings.LastIndex(path, "/"); idx >= 0 {
			parentDir = filepath.Join(basePath, path[:idx+1])
		}

		params["ParentDirectory"] = parentDir
	} else {
		params["DirectoryName"] = basePath

		if basePath == "/" {
			params["DirectoryPrefix"] = ""
		} else {
			params["DirectoryPrefix"] = strings.TrimSuffix(basePath, "/")
		}
	}

	params["Languages"] = parseLanguageHeader(r.Header.Get("Accept-Language"))

	return params
}

// directoryListing creates middleware for handling directory listing in an HTTP file server.
// If enabled, it generates an HTML page showing the directory's contents using a predefined template.
// If it is not enabled, a 403-Forbidden is produced instead of the directory listing.
// The middleware skips directory listing when serving files or paths with index.html present.
func directoryListing(fsys fs.StatFS, enabled bool, basePath string) (func(http.Handler) http.Handler, error) {
	tmpl, err := template.New("directoryListing").Parse(directoryListingTemplate)

	// we accept the downstream nil here. It _must_ work, as it is a core component of SonicWeb's functionality.
	if err != nil {
		return nil, fmt.Errorf("could not parse directory listing template: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// check the desired file is either a directory or an index.html
			path := cleanRequestPath(r.URL.Path)

			if path == "" {
				path = "."
			}

			info, infoErr := fsys.Stat(path)
			hasIndex := false

			// check if index.html is already existing
			if infoErr == nil && info.IsDir() {
				hasIndex = hasIndexFile(fsys, path)
			}

			// jump to the next handler, if we are not to generate the index.html here
			if infoErr != nil || !info.IsDir() || hasIndex {
				next.ServeHTTP(w, r)
				return
			}

			// if we should not generate the index.html, we deny the listing at this point
			if !enabled {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			entries, dirErr := collectDirectoryEntries(fsys, path, basePath)
			if dirErr != nil {
				slog.Error("could not read directory", //nolint:gosec // slog cares for safety
					slog.String("path", cutLog(path)),
					slog.String("error", dirErr.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			params := buildDirectoryListingParams(path, basePath, entries, r)

			outBuf := bytes.Buffer{}

			if err := tmpl.Execute(&outBuf, params); err != nil {
				slog.Error("could not execute directory listing template", //nolint:gosec // slog cares for safety
					slog.String("path", cutLog(path)),
					slog.String("error", err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			if written, err := io.Copy(w, &outBuf); err != nil {
				slog.Error("could not fully send directory listing",
					slog.Int64("written", written),
					slog.String("error", err.Error()))
			}
		})
	}, nil
}
