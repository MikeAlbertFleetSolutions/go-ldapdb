package ldapdb

import (
	"database/sql/driver"
	"fmt"
	"io"
	"strings"

	"gopkg.in/ldap.v2"
)

type resultSet struct {
	results *ldap.SearchResult
	num     int
}

// Rows structure to track results
type Rows struct {
	rc *Conn
	rs resultSet
}

var (
	// errors
	errClosed = fmt.Errorf("rows closed")
)

// Columns returns the columns in the result set
func (rows *Rows) Columns() (data []string) {
	if rows.rs == (resultSet{}) {
		return
	}

	// create slice of all attribute names on first entry
	e := rows.rs.results.Entries[0]
	attrs := make([]string, 0, len(e.Attributes))

	// append attribute names
	for _, attr := range e.Attributes {
		attrs = append(attrs, attr.Name)
	}

	data = attrs
	return
}

// Next navigates to next row in resultset
func (rows *Rows) Next(dest []driver.Value) (err error) {
	if rows.rs == (resultSet{}) {
		err = errClosed
		return
	}

	if rows.rs.num == len(rows.rs.results.Entries) {
		err = io.EOF
		return
	}

	// put attribute values from current entry
	e := rows.rs.results.Entries[rows.rs.num]
	for i, attr := range e.Attributes {
		values := e.GetAttributeValues(attr.Name)
		if len(values) == 1 {
			dest[i] = values[0]
		} else {
			// if multi-value, join with semi-colons
			dest[i] = strings.Join(values, ";")
		}
	}

	rows.rs.num++
	return
}

// Close closes the rows
func (rows *Rows) Close() (err error) {
	if rows.rs == (resultSet{}) {
		err = errClosed
		return
	}
	rows.rc = nil
	rows.rs = resultSet{}

	return
}
