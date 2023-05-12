package ldapdb

import (
	"database/sql/driver"
	"fmt"
	"io"
	"strings"

	"gopkg.in/ldap.v2"
)

type column struct {
	attributeName string
	alias         string
}

// Rows structure to track results
type Rows struct {
	connection *Conn
	columns    []column
	results    *ldap.SearchResult
	currentRow int
}

var (
	// errors
	errClosed = fmt.Errorf("rows closed")
)

// Columns returns the columns in the result set
func (rows *Rows) Columns() []string {
	if rows.results == nil {
		return nil
	}

	// return aliases
	attrs := make([]string, 0, len(rows.columns))
	for _, col := range rows.columns {
		attrs = append(attrs, col.alias)
	}

	return attrs
}

// Next navigates to next row in resultset
func (rows *Rows) Next(dest []driver.Value) error {
	if rows.results == nil {
		return errClosed
	}

	if rows.currentRow == len(rows.results.Entries) {
		return io.EOF
	}

	// put attribute values from current entry
	e := rows.results.Entries[rows.currentRow]
	for i, attr := range rows.columns {
		values := e.GetAttributeValues(attr.attributeName)
		if len(values) == 1 {
			dest[i] = values[0]
		} else {
			// if multi-value, join with semi-colons
			dest[i] = strings.Join(values, ";")
		}
	}

	rows.currentRow++
	return nil
}

// Close closes the rows
func (rows *Rows) Close() error {
	if rows.results == nil {
		return errClosed
	}
	rows.connection = nil
	rows.results = nil

	return nil
}
