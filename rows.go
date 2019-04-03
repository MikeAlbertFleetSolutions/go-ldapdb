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

type resultSet struct {
	results *ldap.SearchResult
	num     int
}

// Rows structure to track results
type Rows struct {
	connection *Conn
	results    resultSet
	columns    []column
}

var (
	// errors
	errClosed = fmt.Errorf("rows closed")
)

// Columns returns the columns in the result set
func (rows *Rows) Columns() (data []string) {
	if rows.results == (resultSet{}) {
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
	if rows.results == (resultSet{}) {
		err = errClosed
		return
	}

	if rows.results.num == len(rows.results.results.Entries) {
		err = io.EOF
		return
	}

	// put attribute values from current entry
	e := rows.results.results.Entries[rows.results.num]
	for i, attr := range rows.columns {
		values := e.GetAttributeValues(attr.attributeName)
		if len(values) == 1 {
			dest[i] = values[0]
		} else {
			// if multi-value, join with semi-colons
			dest[i] = strings.Join(values, ";")
		}
	}

	rows.results.num++
	return
}

// Close closes the rows
func (rows *Rows) Close() (err error) {
	if rows.results == (resultSet{}) {
		err = errClosed
		return
	}
	rows.connection = nil
	rows.results = resultSet{}

	return
}
