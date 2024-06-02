package main

import (
	"os"

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
