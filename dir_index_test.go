package main

import (
	"fmt"
	"io/fs"
	"os"
	"testing"
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
func indexCreateFS(t *testing.T) *os.Root {
	t.Helper()

	dirName := t.TempDir()
	tmpFS, err := os.OpenRoot(dirName)

	if err != nil {
		t.Errorf("could not open temporary root: %v", err)
		return nil
	}

	if err := tmpFS.Mkdir("withIndex", 0o755); err != nil {
		t.Errorf("could not create directory: %v", err)
		return nil
	}

	if err := tmpFS.Mkdir("noIndex", 0o755); err != nil {
		t.Errorf("could not create directory: %v", err)
		return nil
	}

	if err := tmpFS.WriteFile("withIndex/index.html", []byte("index-content"), 0644); err != nil {
		t.Errorf("could not write to withIndex/index.html: %v", err)
		return nil
	}

	if err := tmpFS.WriteFile("noIndex/file.html", []byte("file-content"), 0644); err != nil {
		t.Errorf("could not write to noIndex/file.html: %v", err)
		return nil
	}

	if err := tmpFS.Symlink("noIndex/file.html", "noIndex/link.html"); err != nil {
		t.Errorf("could not create symlink: %v", err)
		return nil
	}

	return tmpFS
}

func TestCollectDirectoryEntries(t *testing.T) {

}
