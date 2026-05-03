// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

// Package dirindex provides functionality for generating directory listings with customizable templates
// and translations.
package dirindex

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/AlphaOne1/sonicweb/utils"
)

//go:embed dir_index.css dir_index.js dir_index.html.tmpl
var directoryListingTemplate embed.FS

// FileEntry represents an entry in a directory, containing its name, file information,
// and symlink target if applicable.
type FileEntry struct {
	Name       string
	Info       fs.FileInfo
	LinkTarget string
}

// renderGen generates a render function for use in HTML templates. It renders the named template into a string
// to utilize pipelined functions on the result.
func renderGen(tmpl *template.Template) func(string, any) (string, error) {
	return func(name string, args any) (string, error) {
		var result bytes.Buffer
		err := tmpl.ExecuteTemplate(&result, name, args)

		return result.String(), err
	}
}

// tindent indents the lines given as argument with the specified number of tabs.
func tindent(numTabs int, lines string) string {
	if numTabs <= 0 {
		return lines
	}

	indent := strings.Repeat("\t", numTabs)
	parts := strings.Split(lines, "\n")

	for i, line := range parts {
		if line != "" {
			parts[i] = indent + line
		}
	}

	return strings.Join(parts, "\n")
}

// cleanRequestPath cleans the URL path by trimming leading and trailing slashes.
func cleanRequestPath(urlPath string) string {
	return strings.Trim(urlPath, "/")
}

// hasIndexFile checks if an index.html file exists in the given directory path.
func hasIndexFile(fsys fs.StatFS, urlPath string) bool {
	indexPath := strings.TrimPrefix(path.Join(urlPath, "index.html"), "./")

	if index, err := fsys.Stat(indexPath); err == nil && !index.IsDir() {
		return true
	}

	return false
}

// processLink resolves the target of a symlink and verifies its existence,
// returning the cleaned link target and indicator of having been able to process it.
func processLink(
	fsys fs.StatFS,
	rawEntry fs.DirEntry,
	urlPath, basePath, absRootPath string) (string, bool) {

	lntgt, err := fs.ReadLink(fsys, path.Join(urlPath, rawEntry.Name()))

	if err != nil {
		return "", false
	}

	lntgt = filepath.ToSlash(lntgt)

	var resolvedTarget string

	if filepath.IsAbs(lntgt) {
		resolvedTarget, err = filepath.Rel(absRootPath, lntgt)

		if err != nil ||
			resolvedTarget == ".." ||
			strings.HasPrefix(resolvedTarget, "../") {
			return "", false
		}
	} else {
		resolvedTarget = path.Join(urlPath, lntgt)
		resolvedTarget = path.Clean(resolvedTarget)
	}

	if _, err := fsys.Stat(resolvedTarget); err != nil {
		return "", false
	}

	var linkTarget string

	if filepath.IsAbs(lntgt) {
		// in case of an absolute link, we directly set the links target instead of the links name
		linkTarget = path.Join(basePath, resolvedTarget)
		linkTarget = "/" + strings.TrimPrefix(linkTarget, "/")
		linkTarget = path.Clean(linkTarget)
	} else {
		// for relative links, we let the file handler resolve it properly
		linkTarget = lntgt
	}

	return linkTarget, true
}

// preCheck handles the initial request validation, directory checks, and language redirection logic for HTTP requests.
func preCheck(
	w http.ResponseWriter,
	r *http.Request,
	fsys fs.StatFS,
	urlPath string,
	next http.Handler,
	enable bool) bool {

	info, infoErr := fsys.Stat(urlPath)
	hasIndex := false

	// check if index.html is already existing
	if infoErr == nil && info.IsDir() {
		hasIndex = hasIndexFile(fsys, urlPath)
	}

	// jump to the next handler, if we are not to generate the index.html here
	if infoErr != nil || !info.IsDir() || hasIndex {
		next.ServeHTTP(w, r)
		return false
	}

	// if we should not generate the index.html, we deny the listing at this point
	if !enable {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return false
	}

	if _, translationFound := Translations[r.URL.Query().Get("lang")]; !translationFound {
		lang, _ := getTranslation(r)

		query := r.URL.Query()
		query.Set("lang", lang)

		redirectPath := "/" + strings.TrimLeft(urlPath, "/")

		if redirectPath == "/." {
			redirectPath = "/"
		}

		redirectPath += "?" + query.Encode()

		w.Header().Add("Vary", "Accept-Language")
		http.Redirect(w, r, redirectPath, http.StatusTemporaryRedirect) //nolint:gosec // G710: same-origin path

		return false
	}

	return true
}

// collectDirectoryEntries reads and processes directory entries, handling symlinks.
func collectDirectoryEntries(fsys fs.StatFS, path, basePath, rootPath string) ([]FileEntry, error) {
	rawEntries, err := fs.ReadDir(fsys, path)

	if err != nil {
		return nil, fmt.Errorf("failed to read directory entries: %w", err)
	}

	absRoot, absRootErr := filepath.Abs(rootPath)

	if absRootErr != nil {
		absRoot = ""
	}

	absRoot = filepath.ToSlash(absRoot)

	entries := make([]FileEntry, 0, len(rawEntries))

	for _, rawEntry := range rawEntries {
		finfo, err := rawEntry.Info()

		if err != nil {
			continue
		}

		linkTarget := ""

		if rawEntry.Type()&fs.ModeSymlink != 0 {
			if l, worked := processLink(fsys, rawEntry, path, basePath, absRoot); worked {
				linkTarget = l
			} else {
				continue
			}
		}

		entries = append(entries, FileEntry{
			Name:       rawEntry.Name(),
			Info:       finfo,
			LinkTarget: linkTarget,
		})
	}

	return entries, nil
}

// getTranslation determines the preferred language from the HTTP request and retrieves the corresponding translation.
// If no match is found, it defaults to "en" and returns the English translation.
func getTranslation(r *http.Request) (string, Translation) {
	requestedLang := r.URL.Query().Get("lang")

	if t, found := Translations[requestedLang]; found {
		return requestedLang, t
	}

	langPrefs := utils.ParseLanguageHeader(r.Header.Get("Accept-Language"))

	for _, l := range langPrefs {
		t, found := Translations[l.Lang]

		if found {
			return l.Lang, t
		}
	}

	// note that we depend on "en" being present
	return "en", Translations["en"]
}

// buildDirectoryListingParams builds the parameters for the directory listing template.
func buildDirectoryListingParams(urlPath, basePath string, entries []FileEntry, r *http.Request) map[string]any {
	params := map[string]any{
		"DirectoryName":   path.Join(basePath, urlPath),
		"DirectoryPrefix": path.Join(basePath, urlPath),
		"Entries":         entries,
	}

	if urlPath != "." {
		parentDir := basePath

		if idx := strings.LastIndex(urlPath, "/"); idx >= 0 {
			// we want the trailing / here, makes clear that it is a directory
			parentDir = path.Join(basePath, urlPath[:idx+1])
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

	params["Language"], params["Translation"] = getTranslation(r)

	return params
}

// DirIndex creates middleware for handling directory listing in an HTTP file server.
// If enabled, it generates an HTML page showing the directory's contents using a predefined template.
// If it is not enabled, a 403-Forbidden is produced instead of the directory listing.
// The middleware skips directory listing when serving files or paths with index.html present.
func DirIndex(fsys fs.StatFS, enable bool, basePath, rootPath string) (func(http.Handler) http.Handler, error) {
	tmpl := template.New("dir_index.html.tmpl")
	tmpl.Funcs(template.FuncMap{
		"render":   renderGen(tmpl),
		"tindent":  tindent,
		"safeJS":   func(s string) template.JS { return template.JS(s) },     //nolint:gosec // we control the input
		"safeCSS":  func(s string) template.CSS { return template.CSS(s) },   //nolint:gosec // we control the input
		"safeHTML": func(s string) template.HTML { return template.HTML(s) }, //nolint:gosec // we control the input
	})
	tmpl, err := tmpl.ParseFS(directoryListingTemplate, "*")

	// we accept the downstream nil here. It _must_ work, as it is a core component of SonicWeb's functionality.
	if err != nil {
		return nil, fmt.Errorf("could not parse directory listing template: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// check the desired file is either a directory or an index.html
			urlPath := cleanRequestPath(r.URL.Path)

			if urlPath == "" {
				urlPath = "."
			}

			if !preCheck(w, r, fsys, urlPath, next, enable) {
				return
			}

			entries, dirErr := collectDirectoryEntries(fsys, urlPath, basePath, rootPath)
			if dirErr != nil {
				slog.Error("could not read directory",
					slog.String("path", utils.CutLog(urlPath)), slog.String("error", dirErr.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			params := buildDirectoryListingParams(urlPath, basePath, entries, r)
			outBuf := bytes.Buffer{}

			if err := tmpl.Execute(&outBuf, params); err != nil {
				slog.Error("could not execute directory listing template",
					slog.String("path", utils.CutLog(urlPath)),
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
