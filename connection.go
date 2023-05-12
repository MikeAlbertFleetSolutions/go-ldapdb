package ldapdb

import (
	"context"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/ldap.v2"
)

// Conn is a ldap connection
type Conn struct {
	pingquery *string
	client    *ldap.Conn
}

var (
	_ driver.Conn = &Conn{}

	// regexs
	rWithWhere      = regexp.MustCompile(`^select\W(.+)\Wfrom\W(.+)\Wwhere\W(.+)$`)
	rWithoutWhere   = regexp.MustCompile(`^select\W(.+)\Wfrom\W(.+)$`)
	rFieldWithAlias = regexp.MustCompile(`^(\w+)\sas\s[\"]?([^"]+)[\"]?$`)

	// errors
	errNoPrepared     = fmt.Errorf("no support for prepared statements")
	errNoTransactions = fmt.Errorf("no support for transactions")
	errBadQuery       = fmt.Errorf("malformed query")
)

// Begin not supported but satisfies the interface requirements
func (c *Conn) Begin() (driver.Tx, error) {
	return nil, errNoTransactions
}

// Prepare not supported but satisfies the interface requirements
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return nil, errNoPrepared
}

// Close closes the ldap connection
func (c *Conn) Close() error {
	c.client.Close()
	return nil
}

// Ping tests the ldap connection
func (c *Conn) Ping(ctx context.Context) error {
	if c.pingquery != nil {
		_, err := c.Query(*c.pingquery, nil)
		return err
	}
	return nil
}

// Query carries out the ldap search
func (c *Conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if len(args) > 0 {
		return nil, errNoPrepared
	}

	//	extract fields, base and filter criteria
	var matches []string
	matches = rWithWhere.FindStringSubmatch(query)
	if len(matches) == 0 {
		matches = rWithoutWhere.FindStringSubmatch(query)
	}

	if len(matches) < 3 {
		return nil, errBadQuery
	}
	fields := strings.Split(matches[1], ",")
	base := matches[2]

	var filter string
	if len(matches) > 3 {
		filter = matches[3]
	} else {
		// a NOP filter, ldap has to have something
		filter = "(objectClass=*)"
	}

	// get list of attributes to search on, figure out any aliases
	// and create column definitions to be used during result processing
	attrs := make([]string, 0, len(fields))
	cols := make([]column, 0, len(fields))
	for _, field := range fields {
		f := strings.TrimSpace(field)
		if f != "" {
			fparts := rFieldWithAlias.FindStringSubmatch(f)

			if len(fparts) > 2 {
				cols = append(cols, column{attributeName: fparts[1], alias: fparts[2]})
				attrs = append(attrs, fparts[1])
			} else {
				cols = append(cols, column{attributeName: f, alias: f})
				attrs = append(attrs, f)
			}
		}
	}

	// form ldap search
	request := ldap.NewSearchRequest(
		base,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		attrs,
		nil,
	)

	// do search
	sr, err := c.client.Search(request)
	if err != nil {
		return nil, err
	}

	rows := &Rows{
		connection: c,
		results:    sr,
		columns:    cols,
	}

	return rows, nil
}
