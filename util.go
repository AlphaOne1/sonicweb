// Copyright the SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"text/template"

	"github.com/google/uuid"
)

// GetEnvDefault gives the content of the given environment variable. If the variable is not
// set or empty, the default given as second argument is instead returned.
func GetEnvDefault(envName, dflt string) string {
	val, isSet := os.LookupEnv(envName)

	if !isSet {
		return dflt
	}

	return val
}

// GetOrCreateID creates a new identifier using uuids. If the given string is already non-empty,
// the same string is returned. In case of an error, the constant string "n/a" is returned.
func GetOrCreateID(id string) string {
	if len(id) > 0 {
		return id
	}

	newID := "n/a"

	if newUuid, err := uuid.NewRandom(); err == nil {
		newID = newUuid.String()
	}

	return newID
}

// ToPointer gives a pointer version of the given value
func ToPointer[T any](v T) *T {
	return &v
}

// PrintLogo takes a text/template string as parameter and renders it to be the logo. It offers the
// following data for the template:
//   - VcsRevision
//   - VcsTime
//   - VcsModified
//   - GoVersion
//
// these can be referenced in the template, e.g. using {{ .VcsRevision }}
func PrintLogo(tmpl string, custom map[string]string) {
	revData := struct {
		VcsRevision string
		VcsTime     string
		VcsModified string
		GoVersion   string
		Values      map[string]string
	}{
		VcsRevision: "unknown",
		VcsTime:     "unknown",
		VcsModified: "",
		GoVersion:   "unknown",
		Values: func() map[string]string {
			if custom != nil {
				return custom
			}
			return make(map[string]string)
		}(),
	}

	if bi, biOK := debug.ReadBuildInfo(); biOK {
		revData.GoVersion = bi.GoVersion

		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				revData.VcsRevision = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					revData.VcsModified = "*"
				}
			case "vcs.time":
				revData.VcsTime = s.Value
			}
		}
	}

	logo := template.New("logo")
	template.Must(logo.Parse(tmpl))

	if err := logo.Execute(os.Stdout, revData); err != nil {
		if jErr := json.NewEncoder(os.Stdout).Encode(revData); jErr != nil {
			slog.Error("logo rendering and version information failed",
				slog.String("logo_error", err.Error()),
				slog.String("info_error", jErr.Error()))
		}
	}

	fmt.Println()
}
