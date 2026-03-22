package main

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
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
	const (
		LangMatch     = 3
		LangPrefMatch = 4
	)

	langHeader = strings.ReplaceAll(langHeader, " ", "")

	langs := make([]LangPref, 0, max(1, len(langHeader)/AverageLangLength))

	for part := range strings.SplitSeq(langHeader, ",") {
		matches := acceptLanguageHeaderRegex.FindStringSubmatch(part)

		switch len(matches) {
		case LangMatch:
			langs = append(langs, LangPref{
				Lang:    matches[2],
				Variant: matches[1],
				Pref:    1.0,
			})
		case LangPrefMatch:
			pref, prefErr := strconv.ParseFloat(matches[3], 32)

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

	slices.SortFunc(langs, func(a, b LangPref) int {
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

	if len(s) > MaxLogStringLength {
		return s[:max(1, MaxLogStringLength-len(EndFill))] + EndFill
	}

	return s
}
