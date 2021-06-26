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
	Integer(name string)

	Timestamps()

	// TODO make dataType an enum?
	MakeColumn(name string, dataType string, flags columnModifier)

	toSQL() string
	hasColumn(name string) bool
}

func (t *table) ID() {
	t.MakeColumn("ID", "CHAR(40)", M_UNIQUE|M_PRIMARY|M_NOT_NULL)
}

func (t *table) Integer(name string) {
	t.MakeColumn(name, "INTEGER", M_NOT_NULL)
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
