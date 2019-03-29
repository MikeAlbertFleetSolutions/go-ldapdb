package ldapdb

import (
	"crypto/tls"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"

	"gopkg.in/ldap.v2"
)

// ldapdriver is a sql.Driver
type ldapdriver struct{}

var (
	_ driver.Driver = &ldapdriver{}

	// regexs
	rConnString = regexp.MustCompile(`^([^/]+)/([^@]+)@([^:]+):(.+)`)

	// errors
	errBadConnString = fmt.Errorf("malformed connection string")
)

func init() {
	sql.Register("ldapdb", &ldapdriver{})
}

// Open prepares the connection to the ldap server
// connString should be of the format: username/password@server:port
func (d ldapdriver) Open(connString string) (c driver.Conn, err error) {
	//	extract properties to use for connection
	matches := rConnString.FindStringSubmatch(connString)
	if len(matches) != 5 {
		err = errBadConnString
		return
	}
	username := matches[1]
	password := matches[2]
	server := matches[3]
	port := matches[4]

	// connect to ldap server
	var l *ldap.Conn
	l, err = ldap.Dial("tcp", fmt.Sprintf("%s:%s", server, port))
	if err != nil {
		return
	}

	// TLS upgrade, #nosec G402
	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return
	}

	// authenticate to ldap server
	err = l.Bind(username, password)
	if err != nil {
		return
	}

	c = &Conn{
		client: l,
	}

	return
}
