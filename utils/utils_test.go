// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package utils_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/AlphaOne1/sonicweb/utils"
)

func TestParseLanguageHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want []utils.LangPref
	}{
		{
			in:   "",
			want: []utils.LangPref{},
		},
		{
			in: "en-US,en;q=0.5",
			want: []utils.LangPref{
				{Lang: "en", Variant: "en-US", Pref: 1},
				{Lang: "en", Variant: "en", Pref: 0.5},
			},
		},
		{
			in: "es;q=0.1,de-DE;q=1,en-US;q=0.5,en;q=0.5",
			want: []utils.LangPref{
				{Lang: "de", Variant: "de-DE", Pref: 1},
				{Lang: "en", Variant: "en-US", Pref: 0.5},
				{Lang: "en", Variant: "en", Pref: 0.5},
				{Lang: "es", Variant: "es", Pref: 0.1},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("TestParseLanguageHeader-%d", i), func(t *testing.T) {
			t.Parallel()

			got := utils.ParseLanguageHeader(test.in)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestCutLog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{
			in:   strings.Repeat("test", 20),
			want: strings.Repeat("test", 15) + "t...",
		},
		{
			in:   "test",
			want: "test",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("TestCutLog-%d", i), func(t *testing.T) {
			t.Parallel()

			got := utils.CutLog(test.in)

			if got != test.want {
				t.Errorf("got %s, want %s", got, test.want)
			}
		})
	}
}
