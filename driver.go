package ldapdb

import (
	"crypto/tls"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"regexp"

	"gopkg.in/ldap.v2"
)

// ldapdriver is a sql.Driver
type ldapdriver struct{}

var (
	_ driver.Driver = &ldapdriver{}

	// regexs
	rConnString = regexp.MustCompile(`^([^/]+)/([^@]+)@([^:]+):([^?]+)(?:\?(.*))?`)

	// errors
	errBadConnString = fmt.Errorf("malformed connection string")
)

func init() {
	sql.Register("ldapdb", &ldapdriver{})
}

// Open prepares the connection to the ldap server
// connString should be of the format: username/password@server:port
func (d ldapdriver) Open(connString string) (driver.Conn, error) {
	//	extract properties to use for connection
	matches := rConnString.FindStringSubmatch(connString)
	if len(matches) != 6 {
		return nil, errBadConnString
	}
	username := matches[1]
	password := matches[2]
	server := matches[3]
	port := matches[4]

	parameters, err := url.ParseQuery(matches[5])
	if err != nil {
		return nil, err
	}

	// connect to ldap server
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%s", server, port))
	if err != nil {
		return nil, err
	}

	// TLS upgrade, #nosec G402
	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}

	// authenticate to ldap server
	err = l.Bind(username, password)
	if err != nil {
		return nil, err
	}

	c := &Conn{
		client: l,
	}

	pingquery := parameters["pingquery"]
	if pingquery != nil {
		c.pingquery = &pingquery[0]
	}

	return c, nil
}
