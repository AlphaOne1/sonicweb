// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package dirindex_test

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlphaOne1/midgard"
	"github.com/AlphaOne1/midgard/defs"
	"github.com/AlphaOne1/midgard/helper"

	"github.com/AlphaOne1/sonicweb/dirindex"
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

			got := dirindex.TcleanRequestPath(test.in)

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
		t.Fatalf("could not open root: %v", fdirErr)
	}

	defer func() { _ = fdir.Close() }()

	statFS, isStatFS := fdir.FS().(fs.StatFS)

	if !isStatFS {
		t.Fatalf("could not get StatFS")
	}

	if dirindex.ThasIndexFile(statFS, ".") {
		t.Errorf("hasIndexFile should return false")
	}

	idx, idxErr := fdir.Create("index.html")

	if idxErr != nil {
		t.Errorf("could not create index.html: %v", idxErr)
	}

	_ = idx.Close()

	if !dirindex.ThasIndexFile(statFS, ".") {
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
		t.Fatalf("could not open temporary root: %v", err)
	}

	if err := tmpFS.Mkdir("withIndex", 0755); err != nil {
		t.Fatalf("could not create directory: %v", err)
	}

	if err := tmpFS.Mkdir("noIndex", 0755); err != nil {
		t.Fatalf("could not create directory: %v", err)
	}

	if err := tmpFS.WriteFile("withIndex/index.html", []byte("index-content"), 0400); err != nil {
		t.Fatalf("could not write to withIndex/index.html: %v", err)
	}

	if err := tmpFS.WriteFile("noIndex/file.html", []byte("file-content"), 0400); err != nil {
		t.Fatalf("could not write to noIndex/file.html: %v", err)
	}

	if err := tmpFS.Symlink("file.html", "noIndex/link.html"); err != nil {
		t.Fatalf("could not create symlink: %v", err)
	}

	if err := tmpFS.Symlink(filepath.Join(dirName, "noIndex/file.html"), "noIndex/abslink.html"); err != nil {
		t.Fatalf("could not create absolute symlink: %v", err)
	}

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.html")

	if err := os.WriteFile(outsideFile, []byte("outside-content"), 0400); err != nil {
		t.Fatalf("could not create file in outside-of-root directory: %v", err)
	}

	if err := tmpFS.Symlink(outsideFile, "noIndex/wrongAbsLink.html"); err != nil {
		t.Fatalf("could not create absolute wrong (points outside of root) symlink: %v", err)
	}

	if err := tmpFS.Symlink("nofile.html", "noIndex/wrongRelLink.html"); err != nil {
		t.Fatalf("could not create relative wrong symlink: %v", err)
	}

	return tmpFS, dirName
}

//nolint:lll //we have to have some longer lines here
func TestCollectDirectoryEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path         string
		indexEnabled bool
		want         []string
		dontWant     []string
		wantStatus   int
	}{
		{
			path:         "/?lang=en",
			indexEnabled: true,
			want: []string{
				"<title> / </title>",
				`<h1> <span class="icon" aria-hidden="true">📂</span> / </h1>`,
				`<td><em><a href="/noIndex"> <span class="icon" aria-hidden="true">📁</span> noIndex/ </a></em></td>`,
				`<td><em><a href="/withIndex"> <span class="icon" aria-hidden="true">📁</span> withIndex/ </a></em></td>`,
			},
			wantStatus: http.StatusOK,
		},
		{
			path:         "/withIndex/",
			indexEnabled: true,
			want: []string{
				"index-content",
			},
			wantStatus: http.StatusOK,
		},
		{
			path:         "/noIndex?lang=en",
			indexEnabled: true,
			want: []string{
				"<title> /noIndex </title>",
				`<h1> <span class="icon" aria-hidden="true">📂</span> /noIndex </h1>`,
				`<td><a href="/noIndex/file.html"> <span class="icon" aria-hidden="true">📄</span> file.html </a></td>`,
				`<td><a href="/noIndex/link.html"> <span class="icon" aria-hidden="true">🔗</span> link.html &rarr; file.html </a></td>`,
				`<td><a href="/noIndex/file.html"> <span class="icon" aria-hidden="true">🔗</span> abslink.html &rarr; /noIndex/file.html </a></td>`,
			},
			dontWant:   []string{"wrongAbsLink.html", "wrongRelLink.html"},
			wantStatus: http.StatusOK,
		},
		{
			path:         "/noIndex/file.html",
			indexEnabled: true,
			want: []string{
				`file-content`,
			},
			wantStatus: http.StatusOK,
		},
		{
			path:         "/noIndex",
			indexEnabled: false,
			wantStatus:   http.StatusForbidden,
		},
		{
			path:         "/withIndex/index.html",
			indexEnabled: true,
			wantStatus:   http.StatusMovedPermanently,
		},
		{
			path:         "/non-existent",
			indexEnabled: true,
			wantStatus:   http.StatusNotFound,
		},
	}

	indexFS, indexDirName := indexCreateFS(t)

	if indexFS == nil {
		t.Fatal("could not create index test filesystem")
	}

	directory, casted := indexFS.FS().(fs.StatFS)

	if !casted {
		t.Fatal("directory does not implement fs.StatFS")
	}

	for testNum, test := range tests {
		t.Run(fmt.Sprintf("TestCollectDirectoryEntries-%d", testNum), func(t *testing.T) {
			t.Parallel()

			testHandler := midgard.StackMiddlewareHandler(
				[]defs.Middleware{
					helper.Must(dirindex.DirIndex(directory, test.indexEnabled, "/", indexDirName)),
				},
				http.FileServerFS(
					directory,
				))

			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, test.path, nil)

			testHandler.ServeHTTP(rec, req)

			if rec.Code != test.wantStatus {
				t.Errorf("got status %d but wanted %d", rec.Code, test.wantStatus)
			}

			for _, want := range test.want {
				if !strings.Contains(rec.Body.String(), want) {
					t.Errorf("expected %q in response body but got %q", want, rec.Body.String())
				}
			}

			for _, dontWant := range test.dontWant {
				if strings.Contains(rec.Body.String(), dontWant) {
					t.Errorf("did not expect %q in response body but got %q", dontWant, rec.Body.String())
				}
			}
		})
	}
}
