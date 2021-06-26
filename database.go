package cservice

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
)

type columnModifier uint8

const (
	M_NOT_NULL = 1 << iota
	M_AUTO_INCREMENT
	M_UNIQUE
	M_PRIMARY
	// M_UNSIGNED

	M_NONE = 0
)

type column struct {
	name          string
	dataType      string
	notNull       bool
	autoIncrement bool
	unique        bool
	primary       bool
}

type table struct {
	name    string
	columns []*column
}

type TableBuilder interface {
	ID()

	// Integer Types
	Tinyint(name string)
	Smallint(name string)
	Mediumint(name string)
	Integer(name string)
	Bigint(name string)

	// TODO
	//
	// From the MySQL documentation:
	// The precision represents the number of significant digits that are stored
	// for values, and the scale represents the number of digits that can be stored
	// following the decimal point. Standard SQL requires that DECIMAL(5,2) be able
	// to store any value with five digits and two decimals, so values that can be
	// stored in the salary column range from -999.99 to 999.99.
	Decimal(name string, precision, scale int)

	// TODO
	//
	// From the MySQL documentation:
	// The precision represents the number of significant digits that are stored
	// for values, and the scale represents the number of digits that can be stored
	// following the decimal point. Standard SQL requires that DECIMAL(5,2) be able
	// to store any value with five digits and two decimals, so values that can be
	// stored in the salary column range from -999.99 to 999.99.
	Numeric(name string, precision, scale int)

	// Floating-point Types
	Float(name string)
	Double(name string)

	// TODO
	//
	// Length can range from 1 to 64 bits.
	Bit(name string, length int)

	Timestamps()

	// TODO make dataType an enum?
	MakeColumn(name string, dataType string, flags columnModifier)

	toSQL() string
	hasColumn(name string) bool
}

func (t *table) ID() {
	t.MakeColumn("ID", "CHAR(40)", M_UNIQUE|M_PRIMARY|M_NOT_NULL)
}

func (t *table) Tinyint(name string) {
	t.MakeColumn(name, "TINYINT", M_NOT_NULL)
}

func (t *table) Smallint(name string) {
	t.MakeColumn(name, "SMALLINT", M_NOT_NULL)
}

func (t *table) Mediumint(name string) {
	t.MakeColumn(name, "MEDIUMINT", M_NOT_NULL)
}

func (t *table) Integer(name string) {
	t.MakeColumn(name, "INTEGER", M_NOT_NULL)
}

func (t *table) Bigint(name string) {
	t.MakeColumn(name, "BIGINT", M_NOT_NULL)
}

func (t *table) Decimal(name string, precision, scale int) {
	t.MakeColumn(name, fmt.Sprintf("DECIMAL(%d, %d)", precision, scale), M_NOT_NULL)
}

func (t *table) Numeric(name string, precision, scale int) {
	t.Decimal(name, precision, scale)
}

func (t *table) Float(name string) {
	t.MakeColumn(name, "FLOAT", M_NOT_NULL)
}

func (t *table) Double(name string) {
	t.MakeColumn(name, "DOUBLE", M_NOT_NULL)
}

func (t *table) Bit(name string, length int) {
	if length < 1 {
		log.Printf("length (%d) passed to Bit column is below the minimum value accepted by this field (1)", length)
		length = 1
	}

	if length > 64 {
		log.Printf("length (%d) passed to Bit column is above the maximum value accepted by this field (64)", length)
		length = 64
	}

	t.MakeColumn(name, fmt.Sprintf("BIT(%d)", length), M_NOT_NULL)
}

func (t *table) Timestamps() {
	t.MakeColumn("CreatedAt", "DATETIME", M_NOT_NULL)
	t.MakeColumn("UpdatedAt", "DATETIME", M_NOT_NULL)
	t.MakeColumn("DeletedAt", "DATETIME", M_NONE)
	// TODO default DeletedAt to NULL
}

func (t *table) MakeColumn(name string, dataType string, flags columnModifier) {
	if t.columns == nil {
		t.columns = []*column{}
	}

	if t.hasColumn(name) {
		log.Printf("column %s already defined in table %s", name, t.name)
		return
	}

	notNull := (flags & M_NOT_NULL) != 0
	autoIncrement := (flags & M_AUTO_INCREMENT) != 0
	unique := (flags & M_UNIQUE) != 0
	primary := (flags & M_PRIMARY) != 0

	t.columns = append(t.columns, &column{
		name:          name,
		dataType:      dataType,
		notNull:       notNull,
		autoIncrement: autoIncrement,
		unique:        unique,
		primary:       primary,
	})
}

func (t *table) toSQL() string {
	var colBuilder strings.Builder
	for _, col := range t.columns {
		var null string = ""
		var autoIncrement string = ""
		var keys string = ""

		if col.notNull {
			null = "NOT NULL "
		}

		if col.autoIncrement {
			autoIncrement = "AUTO_INCREMENT "
		}

		if col.primary || col.unique {
			var primary string = ""
			var unique string = ""

			if col.primary {
				primary = "PRIMARY "
			}

			if col.unique {
				unique = "UNIQUE "
			}

			keys = fmt.Sprintf("%s%sKEY", primary, unique)
		}

		definition := fmt.Sprintf("%s %s %s%s%s,", col.name, col.dataType, null, autoIncrement, keys)
		fmt.Fprint(&colBuilder, definition)
	}

	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(%s)`, t.name, colBuilder.String()[:colBuilder.Len()-1])
}

func (t *table) hasColumn(name string) bool {
	for _, col := range t.columns {
		if col.name == name {
			return true
		}
	}

	return false
}

func BuildTable(tableName string, builder func(TableBuilder)) (string, error) {
	validName, _ := regexp.Match("^[0-9,a-z,A-Z$_]+$", []byte(tableName))
	if !validName {
		return "", fmt.Errorf("table name %s is invalid", tableName)
	}

	tb := &table{
		name: tableName,
	}
	builder(tb)

	if len(tb.columns) == 0 {
		// No columns have been created by the builder function
		return "", errors.New("builder method is empty")
	}

	// Create ID column if it doesn't exist
	tb.ID()

	// Create Timestamps
	tb.Timestamps()

	return tb.toSQL(), nil
}
