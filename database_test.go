package cservice_test

import (
	"bytes"
	"crockerio/cservice"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

// assertHasError checks the expected error was thrown.
//
// An edditional error will be thrown by this method if the expected string is
// empty.
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

// assertStringContains checks the string contains the search string.
//
// An edditional error will be thrown by this method if the expected string is
// empty.
func assertStringContains(t *testing.T, haystack, needle string) {
	if needle == "" {
		// Bail here, otherwise we'll always contain an empty string.
		t.Error("!!test-error!! empty 'needle' string")
	}

	if !strings.Contains(haystack, needle) {
		t.Errorf("expected string (\"%s\") to contain string: \"%s\"", haystack, needle)
	}
}

// assertStringCassertStringMissingontains checks the string does not contain
// the search string.
//
// An edditional error will be thrown by this method if the expected string is
// empty.
func assertStringMissing(t *testing.T, haystack, needle string) {
	if needle == "" {
		// Bail here, otherwise we'll always contain an empty string.
		t.Error("!!test-error!! empty 'needle' string")
	}

	if strings.Contains(haystack, needle) {
		t.Errorf("expected string (\"%s\") to not contain string: \"%s\"", haystack, needle)
	}
}

// TestBuildTable_EmptyBuilder ensures that the BuildTable function returns an
// error when an empty builder function is passed in.
func TestBuildTable_EmptyBuilder(t *testing.T) {
	_, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {})
	assertHasError(t, err, "builder method is empty")
}

// TestBuildTable_AddsGORMColumnsAfterBuilderFunctionRuns encures the BuildTable
// function automatically adds the date columns required by the GORM ORM package
// after successfully running a populated builder function.
//
// These columns are CreatedAt, UpdatedAt and DeletedAt, all of type datetime.
//
// See: https://gorm.io/docs/models.html#gorm-Model
func TestBuildTable_AddsGORMColumnsAfterBuilderFunctionRuns(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.ID()
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "ID CHAR(40) NOT NULL PRIMARY UNIQUE KEY")
	assertStringContains(t, sql, "CreatedAt DATETIME NOT NULL")
	assertStringContains(t, sql, "UpdatedAt DATETIME NOT NULL")
	assertStringContains(t, sql, "DeletedAt DATETIME")
}

// TestBuildTable_OnlyAddsOmittedGORMColumnsAfterBuilderFunctionRuns ensures the
// BuildTable function only adds the GORM columns which haven't already been
// specified within the builder function.
//
// For example, if the builder function specified the CreatedAt column, we don't
// want to recreate that, so only the UpdatedAt and Deleted at columns should be
// added.
func TestBuildTable_OnlyAddsOmittedGORMColumnsAfterBuilderFunctionRuns(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Integer("CreatedAt")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "ID CHAR(40) NOT NULL PRIMARY UNIQUE KEY")
	assertStringContains(t, sql, "CreatedAt INTEGER NOT NULL") // Test that we keep the INTEGER type column created at the start
	assertStringContains(t, sql, "UpdatedAt DATETIME NOT NULL")
	assertStringContains(t, sql, "DeletedAt DATETIME")
}

// TestBuildTable_TableNameValidation ensures the BuildTable method validates
// the given table name to ensure it only contains characters which match the
// following regex: [0-9,a-z,A-Z$_]
//
// If the table name is invalid, an error is returned.
//
// See: https://dev.mysql.com/doc/refman/8.0/en/identifiers.html
func TestBuildTable_TableNameValidation(t *testing.T) {
	testNames := map[string]bool{
		"test":            true,
		"test1234":        true,
		"test_table":      true,
		"TEST":            true,
		"Test1234":        true,
		"Test_Table_1234": true,
		"$Test_1234":      true,
		"Test Table":      false,
		"Test T@ble":      false,
	}

	for name, valid := range testNames {
		t.Run(name, func(t *testing.T) {
			_, err := cservice.BuildTable(name, func(tb cservice.TableBuilder) {
				tb.ID()
			})

			if valid {
				if err != nil {
					t.Errorf("expected %s to be valid, recieved error %s", name, err.Error())
				}
			} else {
				assertHasError(t, err, fmt.Sprintf("table name %s is invalid", name))
			}
		})
	}
}

// TestBuildTable_SkipsColumnIfItAlreadyExists ensures we skip creating a column
// if it has previously been defined and stored within the internal
// table.columns list.
func TestBuildTable_SkipsColumnIfItAlreadyExists(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Integer("Int1")
		tb.Integer("Int1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "Int1 INTEGER")
	assertStringMissing(t, sql, "Int1 INTEGER NOT NULL ,Int1 INTEGER NOT NULL")
}

// TestBuildTable_hasColumn_LogsToTheConsoleIfItFindsDuplicateColumns ensures
// the hasColumn method provides a log message if if encounters a duplicate
// column when building the table.
func TestBuildTable_hasColumn_LogsToTheConsoleIfItFindsDuplicateColumns(t *testing.T) {
	// Capture Logger output.
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
	})

	// Test Below
	_, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Integer("Int1")
		tb.Integer("Int1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, logOutput.String(), "column Int1 already defined in table test")
}

// TODO column types
// TODO flags
