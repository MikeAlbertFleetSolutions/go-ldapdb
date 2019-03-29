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
	client *ldap.Conn
}

var (
	_ driver.Conn    = &Conn{}
	_ driver.Queryer = &Conn{}

	// regexs
	rWithWhere    = regexp.MustCompile(`^select\W(.+)\Wfrom\W(.+)\Wwhere\W(.+)$`)
	rWithoutWhere = regexp.MustCompile(`^select\W(.+)\Wfrom\W(.+)$`)

	// errors
	errNoPrepared     = fmt.Errorf("no support for prepared statements")
	errNoTransactions = fmt.Errorf("no support for transactions")
	errBadQuery       = fmt.Errorf("malformed query")
)

// Begin not supported but satisfies the interface requirements
func (c *Conn) Begin() (tx driver.Tx, err error) {
	err = errNoTransactions
	return
}

// Prepare not supported but satisfies the interface requirements
func (c *Conn) Prepare(query string) (stmt driver.Stmt, err error) {
	err = errNoPrepared
	return
}

// Close closes the ldap connection
func (c *Conn) Close() (err error) {
	c.client.Close()
	return
}

// Ping NOP but satisfies the interface requirements
func (c *Conn) Ping(ctx context.Context) (err error) {
	return
}

// Query carries out the ldap search
func (c *Conn) Query(query string, args []driver.Value) (rows driver.Rows, err error) {
	if len(args) > 0 {
		err = errNoPrepared
		return
	}

	//	extract fields, base and filter criteria
	var matches []string
	matches = rWithWhere.FindStringSubmatch(query)
	if len(matches) == 0 {
		matches = rWithoutWhere.FindStringSubmatch(query)
	}

	if len(matches) < 3 {
		err = errBadQuery
		return
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

	// clean-up attribute names
	attrs := make([]string, 0, len(fields))
	for _, attr := range fields {
		f := strings.TrimSpace(attr)
		if f != "" {
			attrs = append(attrs, f)
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
	var sr *ldap.SearchResult
	sr, err = c.client.Search(request)
	if err != nil {
		return
	}

	rows = &Rows{
		rc: c,
		rs: resultSet{
			results: sr,
		},
	}

	return
}
