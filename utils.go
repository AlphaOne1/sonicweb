// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

type LangPref struct {
	Lang    string
	Variant string
	Pref    float32
}

// acceptLanguageHeaderRegex is a regular expression for parsing the HTTP "Accept-Language" header values
// and quality factors.
var acceptLanguageHeaderRegex = regexp.MustCompile(
	`^(([A-Za-z]+)(?:-[A-Za-z0-9]+)*|\*)(?:;q=(1(?:\.0)?|0(?:\.[0-9]+)?))?$`)

// parseLanguageHeader parses the Accept-Language header and returns a sorted slice of LangPref by preference value.
func parseLanguageHeader(langHeader string) []LangPref {
	const AverageLangLength = 5
	const LangPrefMatch = 4

	langHeader = strings.ReplaceAll(langHeader, " ", "")

	langs := make([]LangPref, 0, max(1, len(langHeader)/AverageLangLength))

	for part := range strings.SplitSeq(langHeader, ",") {
		matches := acceptLanguageHeaderRegex.FindStringSubmatch(part)

		if len(matches) == LangPrefMatch {
			pref, prefErr := strconv.ParseFloat(matches[3], 32)

			// this occurs in the case of an empty match
			if prefErr != nil {
				pref = 1
			}

			langs = append(langs, LangPref{
				Lang:    matches[2],
				Variant: matches[1],
				Pref:    float32(pref),
			})
		}
	}

	slices.SortStableFunc(langs, func(a, b LangPref) int {
		if a.Pref > b.Pref {
			return -1
		} else if a.Pref < b.Pref {
			return 1
		}

		return 0
	})

	return langs
}

// cutLog truncates a string to ensure it does not exceed a specified length, appending a suffix if truncation occurs.
func cutLog(s string) string {
	const MaxLogStringLength = 64 // must be smaller than the MaxPathPartLength!!!
	const EndFill = "..."

	if len([]rune(s)) > MaxLogStringLength {
		return string(append([]rune(s)[:max(1, MaxLogStringLength-len(EndFill))], []rune(EndFill)...))
	}

	return s
}

// executableTime retrieves the modification time of the current executable and
// returns it formatted as an RFC3339 string.
// If the executable lookup fails, the time since the start of the program is returned instead.
func executableTime() string {
	exe, err := os.Executable()

	if err != nil {
		return startTime.Format(time.RFC3339)
	}

	info, err := os.Stat(exe)

	if err != nil {
		return startTime.Format(time.RFC3339)
	}

	return info.ModTime().Format(time.RFC3339)
}
