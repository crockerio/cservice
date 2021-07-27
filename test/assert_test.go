package test_test

import (
	"testing"

	"github.com/crockerio/cservice/test"
)

func TestAssertStringEquals(t *testing.T) {
	test.AssertStringEquals(t, "string", "string")
}
