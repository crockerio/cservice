package cservice_test

import (
	"crockerio/cservice"
	"strings"
	"testing"
)

// TODO
func assertHasError(t *testing.T, out error, expected string) {
	if out == nil {
		t.Error("No error was returned")
	}

	if expected == "" {
		// Bail here, otherwise we'll always contain an empty string.
		t.Error("!!test-error!! empty 'expected' string")
	}

	if !strings.Contains(out.Error(), expected) {
		t.Errorf("expected error (\"%s\") to contain string: \"%s\"", out.Error(), expected)
	}
}

// TestBuildTable_EmptyBuilder ensures that the BuildTable function returns an
// error when an empty builder function is passed in.
func TestBuildTable_EmptyBuilder(t *testing.T) {
	_, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {})
	assertHasError(t, err, "builder method is empty")
}
