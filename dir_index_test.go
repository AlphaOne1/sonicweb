package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/AlphaOne1/midgard"
	"github.com/AlphaOne1/midgard/defs"
	"github.com/AlphaOne1/midgard/helper"
)

func TestCleanRequestPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{
			in:   "foo/",
			want: "foo",
		},

		{
			in:   "/bar",
			want: "bar",
		},
		{
			in:   "/foo/bar/",
			want: "foo/bar",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("TestCleanRequestPath-%d", i), func(t *testing.T) {
			t.Parallel()

			got := cleanRequestPath(test.in)

			if got != test.want {
				t.Errorf("got %s, want %s", got, test.want)
			}
		})
	}
}

func TestHasIndexFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	fdir, fdirErr := os.OpenRoot(dir)

	if fdirErr != nil {
		t.Errorf("could not open root: %v", fdirErr)
	}

	defer func() { _ = fdir.Close() }()

	statFS, isStatFS := fdir.FS().(fs.StatFS)

	if !isStatFS {
		t.Errorf("could not get StatFS")
	}

	if hasIndexFile(statFS, ".") {
		t.Errorf("hasIndexFile should return false")
	}

	idx, idxErr := fdir.Create("index.html")

	if idxErr != nil {
		t.Errorf("could not create index.html: %v", idxErr)
	}

	_ = idx.Close()

	if !hasIndexFile(statFS, ".") {
		t.Errorf("hasIndexFile should return true")
	}
}

// indexCreateFS creates a temporary filesystem structure for testing and returns its root.
// It ensures directories and files with specific structures are created, including symlinks.
//
//	"/"
//	|
//	+--"noIndex"
//	|  |
//	|  +--"file.html"
//	|  '--"link.html"
//	|
//	'--"withIndex"
//	   |
//	   '--"index.html"
func indexCreateFS(t *testing.T) (*os.Root, string) {
	t.Helper()

	dirName := t.TempDir()
	tmpFS, err := os.OpenRoot(dirName)

	if err != nil {
		t.Errorf("could not open temporary root: %v", err)
		return nil, ""
	}

	if err := tmpFS.Mkdir("withIndex", 0o755); err != nil {
		t.Errorf("could not create directory: %v", err)
		return nil, ""
	}

	if err := tmpFS.Mkdir("noIndex", 0o755); err != nil {
		t.Errorf("could not create directory: %v", err)
		return nil, ""
	}

	if err := tmpFS.WriteFile("withIndex/index.html", []byte("index-content"), 0644); err != nil {
		t.Errorf("could not write to withIndex/index.html: %v", err)
		return nil, ""
	}

	if err := tmpFS.WriteFile("noIndex/file.html", []byte("file-content"), 0644); err != nil {
		t.Errorf("could not write to noIndex/file.html: %v", err)
		return nil, ""
	}

	if err := tmpFS.Symlink("file.html", "noIndex/link.html"); err != nil {
		t.Errorf("could not create symlink: %v", err)
		return nil, ""
	}

	if err := tmpFS.Symlink("/noIndex/file.html", "noIndex/abslink.html"); err != nil {
		t.Errorf("could not create absolute symlink: %v", err)
		return nil, ""
	}

	return tmpFS, dirName
}

func TestCollectDirectoryEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path         string
		indexEnabled bool
		want         []string
	}{
		{
			path:         "/",
			indexEnabled: true,
			want: []string{
				"<title> / </title>",
				"<h1> &#128194; / </h1>",
				`<td><a href="/noIndex"> &#128193; noIndex/ </a></td>`,
				`<td><a href="/noIndex"> &#128193; withIndex/ </a></td>`,
			},
		},
		{
			path:         "/withIndex/",
			indexEnabled: true,
			want: []string{
				"index-content",
			},
		},
		{
			path:         "/noIndex",
			indexEnabled: true,
			want: []string{
				"<title> /noIndex </title>",
				"<h1> &#128194; /noIndex </h1>",
				`<td><a href="/noIndex/file.html"> file.html </a></td>`,
				`<td><a href="/noIndex/link.html"> link.html &rarr; file.html </a></td>`,
				`<td><a href="/noIndex/file.html"> abslink.html &rarr; /noIndex/file.html </a></td>`,
			},
		},
	}

	indexFS, indexDirName := indexCreateFS(t)
	directory, casted := indexFS.FS().(fs.StatFS)

	if !casted {
		t.Error("directory does not implement fs.StatFS")
	}

	for testNum, test := range tests {
		t.Run(fmt.Sprintf("TestCollectDirectoryEntries-%d", testNum), func(t *testing.T) {
			t.Parallel()

			testHandler := midgard.StackMiddlewareHandler(
				[]defs.Middleware{
					helper.Must(directoryListing(directory, test.indexEnabled, "/", indexDirName)),
				},
				http.FileServerFS(
					directory,
				))

			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, test.path, nil)

			testHandler.ServeHTTP(rec, req)

			for _, want := range test.want {
				if !strings.Contains(rec.Body.String(), want) {
					t.Errorf("expected %q in response body, got %q", want, rec.Body.String())
				}
			}
		})
	}
}
