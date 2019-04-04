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
func (rows *Rows) Columns() (data []string) {
	if rows.results == nil {
		return
	}

	// return aliases
	attrs := make([]string, 0, len(rows.columns))
	for _, col := range rows.columns {
		attrs = append(attrs, col.alias)
	}

	data = attrs
	return
}

// Next navigates to next row in resultset
func (rows *Rows) Next(dest []driver.Value) (err error) {
	if rows.results == nil {
		err = errClosed
		return
	}

	if rows.currentRow == len(rows.results.Entries) {
		err = io.EOF
		return
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
	return
}

// Close closes the rows
func (rows *Rows) Close() (err error) {
	if rows.results == nil {
		err = errClosed
		return
	}
	rows.connection = nil
	rows.results = nil

	return
}
