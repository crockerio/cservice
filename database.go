package cservice

import "errors"

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
	toSQL() string
}

func (t *table) toSQL() string {
	return ""
}

func BuildTable(tableName string, builder func(TableBuilder)) (string, error) {
	tb := &table{
		name: tableName,
	}
	builder(tb)

	if len(tb.columns) == 0 {
		// No columns have been created by the builder function
		return "", errors.New("builder method is empty")
	}

	return tb.toSQL(), nil
}
