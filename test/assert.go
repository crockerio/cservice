package test

import (
	"strings"
	"testing"
)

func AssertEquals(t *testing.T, expected, recieved interface{}) {
	if expected != recieved {
		t.Errorf("expected %v, got %v", expected, recieved)
	}
}

func AssertStringEquals(t *testing.T, expected, recieved string) {
	if expected != recieved {
		t.Errorf("expected %s, got %s", expected, recieved)
	}
}

func AssertErrorThrown(t *testing.T, err error, expected string) {
	if err == nil {
		t.Errorf("error '%s' not thrown; err is nil", expected)
		return
	}

	if !strings.Contains(err.Error(), expected) {
		t.Errorf("error '%s' does not contain string '%s'", err.Error(), expected)
	}
}
