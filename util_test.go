package main

import (
	"os"
	"testing"
)

func TestGetEnvDefault(t *testing.T) {
	tests := []struct {
		envKey     string
		envDefault string
	}{
		{
			envKey:     "USER",
			envDefault: "test",
		},
		{
			envKey:     "NOTSETANDNEVERWILL",
			envDefault: "test",
		},
	}

	for k, v := range tests {
		got := GetEnvDefault(v.envKey, v.envDefault)
		realEnv, realEnvFound := os.LookupEnv(v.envKey)

		if realEnvFound && got != realEnv {
			t.Errorf("%v: envKey %v is set as %v but wrongly read as %v", k, v.envKey, realEnv, got)
		}

		if !realEnvFound && got != v.envDefault {
			t.Errorf("%v: envKey %v is not set, defaults to %v but wrongly read as %v", k, v.envKey, v.envDefault, got)
		}
	}
}

func TestGetOrCreateID(t *testing.T) {
	tests := []struct {
		in      string
		wantNew bool
	}{
		{
			in:      "",
			wantNew: true,
		},
		{
			in:      "nonsense",
			wantNew: false,
		},
	}

	for k, v := range tests {
		got := GetOrCreateID(v.in)

		if v.wantNew == true && got == v.in {
			t.Errorf("%v: wanted new UUID but got old one", k)
		}

		if !v.wantNew == true && got != v.in {
			t.Errorf("%v: wanted old UUID but got new one", k)
		}
	}
}
