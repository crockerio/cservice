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

// TestBuildTable_DataType_ID ensures the TableBuilder's ID method creates the
// ID column.
//
// The ID column is defined as a 40-length CHAR, which cannot be null and is the
// primary key of the table.
func TestBuildTable_DataType_ID(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.ID()
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "ID CHAR(40) NOT NULL PRIMARY UNIQUE KEY")
}

// TestBuildTable_DataType_Integer ensures the Integer-type columns are created
// correctly.
func TestBuildTable_DataType_Integer(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Integer("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 INTEGER")
}

// TestBuildTable_DataType_Tinyint ensures the Tinyint-type columns are created
// correctly.
func TestBuildTable_DataType_Tinyint(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Tinyint("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 TINYINT")
}

// TestBuildTable_DataType_Smallint ensures the Smallint-type columns are
// created correctly.
func TestBuildTable_DataType_Smallint(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Smallint("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 SMALLINT")
}

// TestBuildTable_DataType_Mediumint ensures the Mediumint-type columns are
// created correctly.
func TestBuildTable_DataType_Mediumint(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Mediumint("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 MEDIUMINT")
}

// TestBuildTable_DataType_Bigint ensures the Bigint-type columns are created
// correctly.
func TestBuildTable_DataType_Bigint(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Bigint("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 BIGINT")
}

// TestBuildTable_DataType_Decimal ensures the Decimal-type columns are created
// correctly.
func TestBuildTable_DataType_Decimal(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Decimal("col1", 5, 2)
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 DECIMAL(5, 2)")
}

// TestBuildTable_DataType_Numeric ensures the Numeric-type columns are created
// correctly.
//
// This is kept as a DECIMAL type as both Decimal and Numeric types are treated
// the same internally in MySQL.
//
// See: https://dev.mysql.com/doc/refman/8.0/en/fixed-point-types.html
func TestBuildTable_DataType_Numeric(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Numeric("col1", 5, 2)
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 DECIMAL(5, 2)")
}

// TestBuildTable_DataType_Float ensures the Float-type columns are created
// correctly.
func TestBuildTable_DataType_Float(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Float("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 FLOAT")
}

// TestBuildTable_DataType_Double ensures the Double-type columns are created
// correctly.
func TestBuildTable_DataType_Double(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Double("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 DOUBLE")
}

// TestBuildTable_DataType_Bit ensures the Bit-type columns are created
// correctly.
func TestBuildTable_DataType_Bit(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Bit("col1", 8)
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 BIT(8)")
}

// TestBuildTable_DataType_Bit_SmallLength ensures the Bit-type method correctly
// handles recieving a bit length which is below the minimum (1) accepted by the
// column.
func TestBuildTable_DataType_Bit_SmallLength(t *testing.T) {
	// Capture Logger output.
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
	})

	// Test Below
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Bit("col1", -1)
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 BIT(1)")
	assertStringContains(t, logOutput.String(), "length (-1) passed to Bit column is below the minimum value accepted by this field (1)")
}

// TestBuildTable_DataType_Bit_LargeLength ensures the Bit-type method correctly
// handles recieving a bit length which is above the maximum (64) accepted by
// the column.
func TestBuildTable_DataType_Bit_LargeLength(t *testing.T) {
	// Capture Logger output.
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
	})

	// Test Below
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Bit("col1", 70)
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 BIT(64)")
	assertStringContains(t, logOutput.String(), "length (70) passed to Bit column is above the maximum value accepted by this field (64)")
}

// TestBuildTable_DataType_Date ensures the Date-type columns are created
// correctly.
func TestBuildTable_DataType_Date(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Date("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 DATE")
}

// TestBuildTable_DataType_DateTime ensures the DateTime-type columns are created
// correctly.
func TestBuildTable_DataType_DateTime(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.DateTime("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 DATETIME")
}

// TestBuildTable_DataType_Timestamp ensures the Timestamp-type columns are
// created correctly.
func TestBuildTable_DataType_Timestamp(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Timestamp("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 TIMESTAMP")
}

// TestBuildTable_DataType_Time ensures the Time-type columns are created
// correctly.
func TestBuildTable_DataType_Time(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Time("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 TIME")
}

// TestBuildTable_DataType_Year ensures the Year-type columns are created
// correctly.
func TestBuildTable_DataType_Year(t *testing.T) {
	sql, err := cservice.BuildTable("test", func(tb cservice.TableBuilder) {
		tb.Year("col1")
	})

	if err != nil {
		t.Errorf("Error thrown: %s", err)
	}

	assertStringContains(t, sql, "col1 YEAR")
}

// TODO flags
